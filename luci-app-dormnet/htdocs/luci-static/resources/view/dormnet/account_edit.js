'use strict';
'require view';
'require form';
'require network';
'require poll';
'require uci';
'require tools.dormnet as dormnet';

function accountId() {
    return L.env.requestpath[5];
}

function allNetworkIds(networks) {
    const ids = [];
    for (const n of (networks || [])) {
        const name = n.getName();
        if (!name || name === 'loopback') continue;
        ids.push(name);
    }
    return ids;
}

// 一个账号 = 一个 bind_iface。section 名固定为 bind_<account>，方便定位与去重。
function ensureBindSection(currentAccountId) {
    const expected = `bind_${currentAccountId}`;
    const owned = uci.sections('dormnet', 'bind_iface').filter(function (s) {
        return (s.parent_account || '') === currentAccountId;
    });

    if (owned.length === 0) {
        // uci.add 返回真实 SID。不同 LuCI 版本对 name 参数支持不同，必须以返回值为准。
        const sid = uci.add('dormnet', 'bind_iface', expected);
        if (!sid) {
            console.error('[dormnet] uci.add returned no SID for', currentAccountId);
            return null;
        }
        uci.set('dormnet', sid, 'parent_account', currentAccountId);
        return sid;
    }

    // 多余的删掉，保留第一个
    const keep = owned[0]['.name'];
    for (let i = 1; i < owned.length; i++) {
        uci.remove('dormnet', owned[i]['.name']);
    }
    return keep;
}

function statusBadge(id) {
    return E('span', {
        id: id,
        style: 'display:inline-block;min-width:6em;padding:.15em .5em;border-radius:.3em;background:#eee;color:#666;font-style:italic;text-align:center;'
    }, _('Not tested'));
}

function setBadge(id, state, text) {
    const el = document.getElementById(id);
    if (!el) return;
    let bg = '#eee', fg = '#666';
    if (state === 'ok')      { bg = '#d4edda'; fg = '#155724'; }
    else if (state === 'bad'){ bg = '#f8d7da'; fg = '#721c24'; }
    else if (state === 'wait'){ bg = '#fff3cd'; fg = '#856404'; }
    el.style.background = bg;
    el.style.color = fg;
    el.style.fontStyle = state ? 'normal' : 'italic';
    el.textContent = text;
}

function setOutput(id, text) {
    const el = document.getElementById(id);
    if (!el) return;
    el.textContent = text || '';
    el.style.display = text ? 'block' : 'none';
}

function runTest(badgeId, outId, iface, tester) {
    if (!iface) {
        setBadge(badgeId, 'bad', _('No interface'));
        setOutput(outId, _('No interface selected.'));
        return Promise.resolve();
    }
    setBadge(badgeId, 'wait', _('Testing…'));
    setOutput(outId, _('Running ping…'));
    return L.resolveDefault(tester(iface)).then(function (r) {
        const ok = r && r.success;
        setBadge(badgeId, ok ? 'ok' : 'bad',
            ok ? _('Connected') : (r && r.message) || _('Unreachable'));
        const data = (r && r.data) || {};
        const header = data.device
            ? `# ping -I ${data.device} ${data.target}\n`
            : `# ${(r && r.message) || ''}\n`;
        setOutput(outId, header + (data.output || ''));
    });
}

function testerCell(prefix, accountId, getIface, tester) {
    const badgeId = `${prefix}_${accountId}`;
    const outId = `${prefix}out_${accountId}`;
    return E('div', { style: 'display:flex;flex-direction:column;gap:.3em;' }, [
        E('div', { style: 'display:flex;gap:.4em;align-items:center;' }, [
            statusBadge(badgeId),
            E('button', {
                'class': 'btn cbi-button cbi-button-action',
                'click': function (ev) {
                    ev.preventDefault();
                    runTest(badgeId, outId, getIface(), tester);
                }
            }, _('Test'))
        ]),
        E('pre', {
            id: outId,
            style: 'display:none;margin:0;padding:.4em .6em;background:#0f111a;color:#d1d5db;font-size:11px;line-height:1.35;border-radius:.3em;max-height:160px;overflow:auto;white-space:pre-wrap;word-break:break-all;'
        }, '')
    ]);
}

// noinspection JSAnnotator
return view.extend({
    load: function () {
        return Promise.all([
            dormnet.supportedTargets(),
            dormnet.extraArgsAccount(accountId()),
            network.getNetworks(),
        ]);
    },
    render: function(data) {
        let m, s, o;

        const supportedTargets = data[0];
        const extraArgs = data[1];
        const networks = data[2] || [];
        const currentAccountId = accountId();

        const rawTargets = supportedTargets && supportedTargets.data;
        const rawExtra   = extraArgs && extraArgs.data;
        const targetList  = Array.isArray(rawTargets) ? rawTargets : [];
        const extraArgList = Array.isArray(rawExtra) ? rawExtra : [];
        if (!Array.isArray(rawTargets)) {
            console.warn('[dormnet] supportedTargets() returned non-array data:', supportedTargets);
        }
        if (!Array.isArray(rawExtra)) {
            console.warn('[dormnet] extraArgsAccount() returned non-array data:', extraArgs);
        }

        // 确保该账号有且仅有一个 bind_iface 段
        const bindSid = ensureBindSection(currentAccountId);

        m = new form.Map('dormnet', `${_('Account Edit')} >> ${currentAccountId}`,
            _('One account binds to exactly one interface; configure it all on this single page.'));

        s = m.section(form.NamedSection, currentAccountId, 'account');

        o = s.option(form.ListValue, 'type', _('Type'));
        o.rmempty = false;
        o.readonly = true;
        for (const target of targetList) {
            o.value(target.id, _(target.name));
        }

        o = s.option(form.Value, 'username', _('Username'));
        o.rmempty = false;

        o = s.option(form.Value, 'password', _('Password'));
        o.password = true;
        o.rmempty = false;

        // ---- 接口 + 运营商等额外参数（实际写入 bind_<account> 段）----
        function bindOption(klass, optName, title) {
            const opt = s.option(klass, `_bind_${optName}`, title);
            opt.cfgvalue = function () {
                return uci.get('dormnet', bindSid, optName) || '';
            };
            opt.write = function (_sid, value) {
                if (value === undefined || value === null || value === '') {
                    uci.unset('dormnet', bindSid, optName);
                } else {
                    uci.set('dormnet', bindSid, optName, value);
                }
            };
            opt.remove = function () {
                uci.unset('dormnet', bindSid, optName);
            };
            return opt;
        }

        const ifaceOpt = bindOption(form.ListValue, 'iface', _('Interface'));
        ifaceOpt.rmempty = false;
        for (const id of allNetworkIds(networks)) {
            ifaceOpt.value(id, id);
        }

        for (const arg of extraArgList) {
            if (!arg || !arg.type || !form[arg.type]) continue;
            o = bindOption(form[arg.type], arg.id, _(arg.title));
            o.password = !!arg.is_pwd;
            o.description = arg.desc;
            if (arg.default !== undefined) o.default = arg.default;
            if (arg.required) {
                o.rmempty = false;
            } else {
                o.optional = true;
            }
            if (arg.type === 'ListValue' && Array.isArray(arg.candidates)) {
                for (const item of arg.candidates) {
                    o.value(item.value, _(item.name));
                }
            }
        }

        // ---- 连通性测试 ----
        o = s.option(form.Value, 'connectivity_check_interval', _('Auto check interval (s)'));
        o.description = _('Seconds between automatic connectivity checks. 0 disables auto check.');
        o.datatype = 'uinteger';
        o.default = '0';
        o.placeholder = '0';

        function currentIface() {
            // 优先取表单当前值（用户刚选完还没保存也能测）
            try {
                const v = ifaceOpt.formvalue(currentAccountId);
                if (v) return v;
            } catch (e) {}
            return uci.get('dormnet', bindSid, 'iface') || '';
        }

        o = s.option(form.DummyValue, '_internet', _('Internet'));
        o.cfgvalue = function () {
            return testerCell('internet', currentAccountId, currentIface, dormnet.pingInternet);
        };

        o = s.option(form.DummyValue, '_campus', _('Campus'));
        o.cfgvalue = function () {
            return testerCell('campus', currentAccountId, currentIface, dormnet.pingCampus);
        };

        const checkInterval = parseInt(uci.get('dormnet', currentAccountId, 'connectivity_check_interval') || '0', 10);
        if (checkInterval) {
            poll.add(function () {
                const iface = currentIface();
                return Promise.all([
                    runTest(`internet_${currentAccountId}`, `internetout_${currentAccountId}`, iface, dormnet.pingInternet),
                    runTest(`campus_${currentAccountId}`, `campusout_${currentAccountId}`, iface, dormnet.pingCampus),
                ]);
            }, checkInterval);
        }

        return m.render();
    }
});
