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

function statusBadge(id, label) {
    return E('span', {
        id: id,
        style: 'display:inline-block;min-width:5em;padding:.15em .5em;border-radius:.3em;background:#eee;color:#666;font-style:italic;text-align:center;'
    }, label || _('Not tested'));
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

function runTest(badgeId, iface, tester) {
    if (!iface) {
        setBadge(badgeId, 'bad', _('No interface'));
        return Promise.resolve();
    }
    setBadge(badgeId, 'wait', _('Testing…'));
    return L.resolveDefault(tester(iface)).then(function (r) {
        const ok = r && r.success;
        setBadge(badgeId, ok ? 'ok' : 'bad', ok ? _('Connected') : _('Unreachable'));
    });
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

        const targetList = (supportedTargets && supportedTargets.data) || [];
        const extraArgList = (extraArgs && extraArgs.data) || [];

        m = new form.Map('dormnet', `${_('Account Edit')} >> ${currentAccountId}`);

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

        o = s.option(form.Value, 'connectivity_check_interval', _('Connectivity check interval'));
        o.description = _('Seconds between automatic connectivity checks. Set to 0 to disable automatic checks.');
        o.datatype = 'uinteger';
        o.default = '0';
        o.placeholder = '0';

        s = m.section(form.GridSection, 'bind_iface', _('Bound Interfaces'),
            _('Each row represents one interface used both to bind and to perform campus network login.'));
        s.anonymous = true;
        s.addremove = true;
        s.sortable = true;
        s.cloneable = true;
        s.nodescriptions = true;
        s.filter = function (section_id) {
            const account = uci.get('dormnet', section_id, 'parent_account') || '';
            return account === currentAccountId;
        };

        o = s.option(form.HiddenValue, 'parent_account');
        o.default = currentAccountId;
        o.write = function (section_id) {
            return uci.set('dormnet', section_id, 'parent_account', currentAccountId);
        };

        o = s.option(form.ListValue, 'iface', _('Interface'));
        o.rmempty = false;
        for (const id of allNetworkIds(networks)) {
            o.value(id, id);
        }

        for (const arg of extraArgList) {
            if (!arg || !arg.type || !form[arg.type]) continue;
            o = s.option(form[arg.type], arg.id, _(arg.title));
            o.password = !!arg.is_pwd;
            o.description = arg.desc;
            if (arg.default !== undefined) o.default = arg.default;
            if (arg.required) {
                o.rmempty = false;
            } else {
                o.optional = true;
            }
            if (arg.modalonly) o.modalonly = true;
            if (arg.type === 'ListValue' && Array.isArray(arg.candidates)) {
                for (const item of arg.candidates) {
                    o.value(item.value, _(item.name));
                }
            }
        }

        // 公网状态 + 测试按钮
        o = s.option(form.DummyValue, '_internet', _('Internet'));
        o.modalonly = false;
        o.cfgvalue = function (section_id) {
            const iface = uci.get('dormnet', section_id, 'iface') || '';
            return E('div', { style: 'display:flex;gap:.4em;align-items:center;' }, [
                statusBadge(`internet_${section_id}`),
                E('button', {
                    'class': 'btn cbi-button cbi-button-action',
                    'click': function (ev) {
                        ev.preventDefault();
                        runTest(`internet_${section_id}`, iface, dormnet.pingInternet);
                    }
                }, _('Test'))
            ]);
        };

        // 校园网状态 + 测试按钮
        o = s.option(form.DummyValue, '_campus', _('Campus'));
        o.modalonly = false;
        o.cfgvalue = function (section_id) {
            const iface = uci.get('dormnet', section_id, 'iface') || '';
            return E('div', { style: 'display:flex;gap:.4em;align-items:center;' }, [
                statusBadge(`campus_${section_id}`),
                E('button', {
                    'class': 'btn cbi-button cbi-button-action',
                    'click': function (ev) {
                        ev.preventDefault();
                        runTest(`campus_${section_id}`, iface, dormnet.pingCampus);
                    }
                }, _('Test'))
            ]);
        };

        const checkInterval = parseInt(uci.get('dormnet', currentAccountId, 'connectivity_check_interval') || '0', 10);
        if (checkInterval) {
            poll.add(function () {
                const tasks = [];
                for (const sid of uci.sections('dormnet', 'bind_iface').map(s => s['.name'])) {
                    if ((uci.get('dormnet', sid, 'parent_account') || '') !== currentAccountId) continue;
                    const iface = uci.get('dormnet', sid, 'iface') || '';
                    tasks.push(runTest(`internet_${sid}`, iface, dormnet.pingInternet));
                    tasks.push(runTest(`campus_${sid}`, iface, dormnet.pingCampus));
                }
                return Promise.all(tasks);
            }, checkInterval);
        }

        return m.render();
    }
});
