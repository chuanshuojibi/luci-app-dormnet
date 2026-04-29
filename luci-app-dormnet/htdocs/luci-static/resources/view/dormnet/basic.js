'use strict';
'require view';
'require form';
'require poll';
'require ui';
'require uci';
'require tools.dormnet as dormnet';

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
        setBadge(badgeId, 'bad', _('No iface'));
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

function loginBadgeColors(state) {
    switch (state) {
        case 'Online':    return { bg: '#d4edda', fg: '#155724' };
        case 'LoggingIn': return { bg: '#cfe2ff', fg: '#084298' };
        case 'Checking':  return { bg: '#fff3cd', fg: '#856404' };
        case 'Backoff':   return { bg: '#fff3cd', fg: '#856404' };
        case 'Failed':    return { bg: '#f8d7da', fg: '#721c24' };
        case 'GivenUp':   return { bg: '#f5c6cb', fg: '#721c24' };
        default:          return { bg: '#eee',    fg: '#666'    };
    }
}

function loginBadgeText(st) {
    if (!st) return _('Unknown');
    if (st.status === 'Backoff' && st.next_attempt_at) {
        const remain = Math.max(0, st.next_attempt_at - Math.floor(Date.now() / 1000));
        const mm = Math.floor(remain / 60), ss = remain % 60;
        return `${st.status} (#${st.retry||0}, ${mm}:${ss < 10 ? '0' + ss : ss})`;
    }
    if (st.status === 'GivenUp') return `GivenUp (#${st.retry || 0})`;
    if (st.status === 'Failed' && st.last_error) return 'Failed';
    return st.status || _('Unknown');
}

function renderLoginBadge(id, st) {
    const c = loginBadgeColors(st && st.status);
    return E('span', {
        id: id,
        title: (st && st.last_error) || '',
        style: 'display:inline-block;min-width:7em;padding:.15em .5em;border-radius:.3em;background:' + c.bg + ';color:' + c.fg + ';text-align:center;font-size:12px;'
    }, loginBadgeText(st));
}

function applyLoginBadge(id, st) {
    const el = document.getElementById(id);
    if (!el) return;
    const c = loginBadgeColors(st && st.status);
    el.style.background = c.bg;
    el.style.color = c.fg;
    el.title = (st && st.last_error) || '';
    el.textContent = loginBadgeText(st);
}

function renderStatus(running) {
    return updateStatus(E('input', { id: 'running_status', style: 'border: unset; font-style: italic; font-weight: bold;', readonly: '' }), running);
}

function updateStatus(element, running) {
    if (element) {
        element.style.color = running ? 'green' : 'red';
        element.value = running ? _('Running') : _('Not Running');
    }
    return element;
}

// noinspection JSAnnotator
return view.extend({
    load: function () {
        return Promise.all([
            dormnet.status(),
            dormnet.buildInfo(),
            uci.load('dormnet'),
            L.resolveDefault(dormnet.accountStatus(), { success: false, data: { states: [] } }),
        ]);
    },
    render: function(data) {
        const status = data[0];
        const buildInfo = data[1];
        const initStatus = data[3] || {};
        const initStates = ((initStatus.data || {}).states) || [];
        const stateMap = {};
        for (const st of initStates) {
            stateMap[`${st.account}|${st.iface}`] = st;
        }

        let m, s, o;

        m = new form.Map('dormnet', `${_('DormNet')} v${buildInfo.data.version}`,
            `${_('Manage your campus network with DormNet on OpenWrt.')}`);

        s = m.section(form.NamedSection, 'basic', 'basic');

        o = s.option(form.Flag, 'enabled', _('Enabled'));
        o.default = '0';

        o = s.option(form.Button, 'restart', _('Restart'));
        o.description = _("Restart manually");
        o.depends('enabled', '1');
        o.inputtitle = _('Restart');
        o.inputstyle = 'apply';
        o.onclick = function () {
            return L.resolveDefault(dormnet.restart()).then(function () {
                ui.addNotification(null, E('p', _('DormNet restarted.')), 'info');
            });
        };

        o = s.option(form.DummyValue, '_running_status', _('Running status'));
        o.cfgvalue = function () {
            return renderStatus(status.data.running);
        };
        poll.add(function () {
            return L.resolveDefault(dormnet.status()).then(function (status) {
                updateStatus(document.getElementById('running_status'), status.data.running);
            });
        });

        o = s.option(form.ListValue, 'work_mode', _('Work mode'));
        o.default = 'master';
        o.value('master', _('Master Mode'));
        o.value('peer', _('Peer Mode'));

        // 多选：为空时 = 所有账号都参与
        o = s.option(form.MultiValue, 'login_account', _('Login accounts'));
        o.description = _('Select one or more saved accounts used for campus network login. Leave empty to use all.');
        o.optional = true;
        o.display_size = 4;
        for (const account of uci.sections('dormnet', 'account')) {
            o.value(account['.name'], account.username || account['.name']);
        }

        o = s.option(form.Flag, 'use_sd_network', _('Use software-defined networking'));
        o.description = _("Select when you need multiple geographically dispersed devices to use the campus network simultaneously.")
        o.default = '0';

        o = s.option(form.ListValue, 'work_with', _('Work with'));
        o.default = 'easytier';
        o.depends('use_sd_network', '1');
        o.value('easytier', _('EasyTier'));
        // o.value('zerotier', _('ZeroTier'));

        o = s.option(form.Value, 'listen', _('Listen address'));
        o.depends({
            'work_mode': 'master',
            'work_with': 'easytier'
        });
        o.default = "127.0.0.1:10721";

        // ============ 连通性总览 ============
        const accounts = uci.sections('dormnet', 'account');
        const binds = uci.sections('dormnet', 'bind_iface');

        // 按 parent_account 分组的 bind_iface
        const byAccount = {};
        for (const b of binds) {
            const pa = b.parent_account || '';
            (byAccount[pa] = byAccount[pa] || []).push(b);
        }

        const overviewRows = [];
        for (const acc of accounts) {
            const sid = acc['.name'];
            const list = byAccount[sid] || [];
            if (list.length === 0) {
                overviewRows.push(E('tr', {}, [
                    E('td', {}, acc.username || sid),
                    E('td', { colspan: 5, style: 'color:#888;font-style:italic;' },
                        _('No bound interfaces. Configure in Account Setting.'))
                ]));
                continue;
            }
            for (const b of list) {
                const ifname = b.iface || '';
                const bsid = b['.name'];
                const tag = `${sid}_${bsid}`;
                const loginBadgeId = `ov_login_${tag}`;
                const stKey = `${sid}|${ifname}`;
                overviewRows.push(E('tr', { 'data-state-key': stKey }, [
                    E('td', {}, acc.username || sid),
                    E('td', {}, ifname || E('em', _('unset'))),
                    E('td', {}, renderLoginBadge(loginBadgeId, stateMap[stKey])),
                    E('td', {}, [
                        E('div', { style: 'display:flex;flex-direction:column;gap:.3em;' }, [
                            E('div', { style: 'display:flex;gap:.4em;align-items:center;' }, [
                                statusBadge(`ov_internet_${tag}`),
                                E('button', {
                                    'class': 'btn cbi-button cbi-button-action',
                                    'click': function (ev) {
                                        ev.preventDefault();
                                        runTest(`ov_internet_${tag}`, `ov_internetout_${tag}`, ifname, dormnet.pingInternet);
                                    }
                                }, _('Test'))
                            ]),
                            E('pre', {
                                id: `ov_internetout_${tag}`,
                                style: 'display:none;margin:0;padding:.4em .6em;background:#0f111a;color:#d1d5db;font-size:11px;line-height:1.35;border-radius:.3em;max-height:140px;overflow:auto;white-space:pre-wrap;'
                            }, '')
                        ])
                    ]),
                    E('td', {}, [
                        E('div', { style: 'display:flex;flex-direction:column;gap:.3em;' }, [
                            E('div', { style: 'display:flex;gap:.4em;align-items:center;' }, [
                                statusBadge(`ov_campus_${tag}`),
                                E('button', {
                                    'class': 'btn cbi-button cbi-button-action',
                                    'click': function (ev) {
                                        ev.preventDefault();
                                        runTest(`ov_campus_${tag}`, `ov_campusout_${tag}`, ifname, dormnet.pingCampus);
                                    }
                                }, _('Test'))
                            ]),
                            E('pre', {
                                id: `ov_campusout_${tag}`,
                                style: 'display:none;margin:0;padding:.4em .6em;background:#0f111a;color:#d1d5db;font-size:11px;line-height:1.35;border-radius:.3em;max-height:140px;overflow:auto;white-space:pre-wrap;'
                            }, '')
                        ])
                    ]),
                ]));
            }
        }

        const testAllBtn = E('button', {
            'class': 'btn cbi-button cbi-button-action',
            'style': 'margin:.5em 0;',
            'click': function (ev) {
                ev.preventDefault();
                const tasks = [];
                for (const acc of accounts) {
                    const list = byAccount[acc['.name']] || [];
                    for (const b of list) {
                        const tag = `${acc['.name']}_${b['.name']}`;
                        tasks.push(runTest(`ov_internet_${tag}`, `ov_internetout_${tag}`, b.iface || '', dormnet.pingInternet));
                        tasks.push(runTest(`ov_campus_${tag}`, `ov_campusout_${tag}`, b.iface || '', dormnet.pingCampus));
                    }
                }
                return Promise.all(tasks);
            }
        }, _('Test all'));

        const overview = E('div', { 'class': 'cbi-section', 'style': 'margin-top:1em;' }, [
            E('h3', {}, _('Connectivity Overview')),
            E('div', { 'class': 'cbi-section-descr' },
                _('Quick connectivity test for every bound interface across all accounts.')),
            testAllBtn,
            accounts.length === 0
                ? E('div', { style: 'color:#888;font-style:italic;' },
                    _('No accounts yet. Add one in Account Setting.'))
                : E('table', { 'class': 'table cbi-section-table', 'style': 'width:100%;' }, [
                    E('thead', {}, E('tr', {}, [
                        E('th', { 'class': 'th' }, _('Account')),
                        E('th', { 'class': 'th' }, _('Interface')),
                        E('th', { 'class': 'th' }, _('Login Status')),
                        E('th', { 'class': 'th' }, _('Internet')),
                        E('th', { 'class': 'th' }, _('Campus')),
                    ])),
                    E('tbody', {}, overviewRows),
                ]),
        ]);

        // 5 \u79d2\u8f6e\u8be2\u4e00\u6b21\u767b\u5f55\u72b6\u6001\uff0c\u5237\u65b0\u5f92\u7ae0
        poll.add(function () {
            return L.resolveDefault(dormnet.accountStatus(),
                { success: false, data: { states: [] } }).then(function (resp) {
                const sts = ((resp && resp.data) || {}).states || [];
                const m = {};
                for (const st of sts) {
                    m[`${st.account}|${st.iface}`] = st;
                }
                document.querySelectorAll('tr[data-state-key]').forEach(function (tr) {
                    const k = tr.getAttribute('data-state-key');
                    const idMatch = tr.querySelector('span[id^="ov_login_"]');
                    if (idMatch) applyLoginBadge(idMatch.id, m[k]);
                });
            });
        }, 5);

        return Promise.resolve(m.render()).then(function (mapNode) {
            return E('div', {}, [mapNode, overview]);
        });
    }
});
