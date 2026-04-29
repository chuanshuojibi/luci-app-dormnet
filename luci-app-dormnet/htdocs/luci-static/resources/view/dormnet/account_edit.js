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

function wanNetworkIds() {
    const ids = [];

    for (const zone of uci.sections('firewall', 'zone')) {
        if (zone.name !== 'wan') {
            continue;
        }

        const networks = L.toArray(zone.network);
        for (const network of networks) {
            ids.push(network);
        }
    }

    return ids;
}

function selectedLoginIface(currentAccountId) {
    return uci.get('dormnet', currentAccountId, 'login_iface') || wanNetworkIds()[0] || '';
}

function renderConnectivityStatus(id) {
    return E('span', { id: id }, _('Not tested'));
}

function updateConnectivityStatus(id, result) {
    const element = document.getElementById(id);
    if (!element) {
        return;
    }

    element.style.color = result.success ? 'green' : 'red';
    element.textContent = result.success ? _('Connected') : _('Unreachable');
}

function testConnectivity(currentAccountId, tester, statusId) {
    const iface = selectedLoginIface(currentAccountId);
    if (!iface) {
        updateConnectivityStatus(statusId, { success: false });
        return Promise.resolve();
    }

    return L.resolveDefault(tester(iface)).then(function (result) {
        updateConnectivityStatus(statusId, result);
    });
}

// noinspection JSAnnotator
return view.extend({
    load: function () {
        return Promise.all([
            dormnet.supportedTargets(),
            dormnet.extraArgsAccount(accountId()),
            network.getNetworks(),
            uci.load('firewall'),
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

        o = s.option(form.ListValue, 'login_iface', _('Login interface'));
        o.description = _('The WAN interface used to provide IP and MAC address for campus network login.');
        o.optional = true;
        o.value('', _('Auto'));
        for (const networkId of wanNetworkIds()) {
            o.value(networkId, networkId);
        }

        o = s.option(form.Value, 'connectivity_check_interval', _('Connectivity check interval'));
        o.description = _('Seconds between automatic connectivity checks. Set to 0 to disable automatic checks.');
        o.datatype = 'uinteger';
        o.default = '0';
        o.placeholder = '0';

        o = s.option(form.Button, '_test_internet', _('Test internet connectivity'));
        o.inputtitle = _('Test');
        o.inputstyle = 'apply';
        o.onclick = function () {
            return testConnectivity(currentAccountId, dormnet.pingInternet, 'internet_status');
        };

        o = s.option(form.DummyValue, '_internet_status', _('Internet connectivity'));
        o.cfgvalue = function () {
            return renderConnectivityStatus('internet_status');
        };

        o = s.option(form.Button, '_test_campus', _('Test campus connectivity'));
        o.inputtitle = _('Test');
        o.inputstyle = 'apply';
        o.onclick = function () {
            return testConnectivity(currentAccountId, dormnet.pingCampus, 'campus_status');
        };

        o = s.option(form.DummyValue, '_campus_status', _('Campus connectivity'));
        o.cfgvalue = function () {
            return renderConnectivityStatus('campus_status');
        };

        const connectivityCheckInterval = parseInt(uci.get('dormnet', currentAccountId, 'connectivity_check_interval') || '0', 10);
        if (connectivityCheckInterval) {
            poll.add(function () {
                return Promise.all([
                    testConnectivity(currentAccountId, dormnet.pingInternet, 'internet_status'),
                    testConnectivity(currentAccountId, dormnet.pingCampus, 'campus_status'),
                ]);
            }, connectivityCheckInterval);
        }

        s = m.section(form.GridSection, 'bind_iface', _('Bound Interfaces'));
        s.anonymous = true;
        s.addremove = true;
        s.sortable = true;
        s.cloneable = true;
        s.nodescriptions = true;
        s.filter = function (section_id) {
            const account = uci.get('dormnet', section_id, 'parent_account') || '';
            return account === currentAccountId;
        };

        // parent_account 必须在内联添加场景下也写入，否则 filter 会立即把新行过滤掉。
        // 不设 modalonly：HiddenValue 本身就不可见，但 parse() 会跑、default 会写。
        o = s.option(form.HiddenValue, 'parent_account');
        o.default = currentAccountId;
        o.write = function (section_id) {
            return uci.set('dormnet', section_id, 'parent_account', currentAccountId);
        };

        o = s.option(form.ListValue, 'iface', _('Interface'));
        o.rmempty = false;
        for (const network of networks) {
            const name = network.getName();
            if (!name || name === 'loopback') {
                continue;
            }
            o.value(name, name);
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

        return m.render();
    }
});
