'use strict';
'require view';
'require form';
'require uci';
'require tools.dormnet as dormnet';

// noinspection JSAnnotator
return view.extend({
    load: function () {
        return Promise.all([
            dormnet.supportedTargets(),
        ]);
    },
    render: function(data) {
        const supportedTargets = data[0];

        let m, s, o;

        m = new form.Map('dormnet', _('Account Setting'),
            _('You can add one or more accounts for logging into the campus network.'));

        s = m.section(form.GridSection, 'account', _('Account List'));
        s.addremove = true;
        s.sortable = true;
        s.extedit = '/cgi-bin/luci/admin/services/dormnet/account/edit/%s';
        s.modaltitle = _('Edit account');
        s.renderRowActions = function(section_id) {
            return this.super('renderRowActions', [ section_id ]);
        }

        o = s.option(form.ListValue, 'type', _('Type'));
        o.rmempty = false;
        for (const target of supportedTargets.data) {
            o.value(target.id, _(target.name));
        }

        o = s.option(form.Value, 'username', _('Username'));
        o.rmempty = false;

        o = s.option(form.Value, 'password', _('Password'));
        o.password = true;
        o.modalonly = true;
        o.rmempty = false;

        o = s.option(form.DummyValue, 'login_iface', _('Login interface'));
        o.cfgvalue = function (section_id) {
            return uci.get('dormnet', section_id, 'login_iface') || _('Auto');
        };

        o = s.option(form.Value, '_iface_conf_count', _('Bound Interfaces'));
        o.readonly = true;
        o.modalonly = false;
        o.cfgvalue = function (section_id) {
            const list = uci.sections('dormnet', 'bind_iface');
            const ifaces = [];
            for (const item of list) {
                if (item.parent_account === section_id) {
                    ifaces.push(item.iface);
                }
            }
            return ifaces.length > 0 ? ifaces.join(', ') : _('Empty');
        };

        return m.render();
    }
});
