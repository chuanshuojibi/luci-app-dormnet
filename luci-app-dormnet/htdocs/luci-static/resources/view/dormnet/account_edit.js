'use strict';
'require view';
'require form';
'require network';
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
        const networks = data[2];
        const currentAccountId = accountId();

        m = new form.Map('dormnet', `${_('Account Edit')} >> ${currentAccountId}`);

        s = m.section(form.NamedSection, currentAccountId, 'account');

        o = s.option(form.ListValue, 'type', _('Type'));
        o.rmempty = false;
        o.readonly = true;
        for (const target of supportedTargets.data) {
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

        o = s.option(form.HiddenValue, 'parent_account');
        o.modalonly = true;
        o.default = currentAccountId;

        o = s.option(form.ListValue, 'iface', _('Interface'));
        for (const network of networks) {
            if (network.sid === 'loopback') {
                continue;
            }
            o.value(network.sid, network.sid);
        }

        for (const arg of extraArgs.data) {
            o = s.option(form[arg.type], arg.id, _(arg.title));
            o.password = arg.is_pwd;
            o.description = arg.desc;
            o.default = arg.default;
            if (arg.required) {
                o.rmempty = false;
            } else {
                o.optional = true;
            }
            if (arg.modalonly) {
                o.modalonly = true;
            }
            if (arg.type === 'ListValue') {
                for (const item of arg.candidates) {
                    o.value(item.value, _(item.name))
                }
            }
        }

        return m.render();
    }
});
