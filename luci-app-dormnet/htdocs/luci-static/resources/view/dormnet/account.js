'use strict';
'require view';
'require form';
'require network';
'require uci';
'require tools.dormnet as dormnet';

function wanNetworkIds() {
    const ids = [];
    for (const zone of uci.sections('firewall', 'zone')) {
        if (zone.name !== 'wan') continue;
        for (const n of L.toArray(zone.network)) ids.push(n);
    }
    return ids;
}

// noinspection JSAnnotator
return view.extend({
    load: function () {
        return Promise.all([
            dormnet.supportedTargets(),
            network.getNetworks(),
            uci.load('firewall'),
        ]);
    },
    render: function(data) {
        const supportedTargets = data[0];
        const targetList = (supportedTargets && supportedTargets.data) || [];

        let m, s, o;

        m = new form.Map('dormnet', _('Account Setting'),
            _('You can add one or more accounts for logging into the campus network.'));

        // 用 TableSection：无 modal，所有字段直接在表格里编辑，避免“假编辑”错觉
        s = m.section(form.TableSection, 'account', _('Account List'));
        s.addremove = true;
        s.sortable = true;
        s.anonymous = false;
        s.addbtntitle = _('Add account');

        o = s.option(form.ListValue, 'type', _('Type'));
        o.rmempty = false;
        for (const target of targetList) {
            o.value(target.id, _(target.name));
        }

        o = s.option(form.Value, 'username', _('Username'));
        o.rmempty = false;

        o = s.option(form.Value, 'password', _('Password'));
        o.password = true;
        o.rmempty = false;

        o = s.option(form.ListValue, 'login_iface', _('Login interface'));
        o.value('', _('Auto'));
        for (const id of wanNetworkIds()) {
            o.value(id, id);
        }

        o = s.option(form.DummyValue, '_iface_conf_count', _('Bound Interfaces'));
        o.cfgvalue = function (section_id) {
            const ifaces = [];
            for (const item of uci.sections('dormnet', 'bind_iface')) {
                if (item.parent_account === section_id) {
                    ifaces.push(item.iface);
                }
            }
            return ifaces.length > 0 ? ifaces.join(', ') : E('em', _('None'));
        };

        // 明确的“配置绑定接口”跳转按钮，跳到 account_edit 页面
        o = s.option(form.Button, '_advanced', _('Configure'));
        o.inputtitle = _('Bind interfaces');
        o.inputstyle = 'edit';
        o.onclick = function (ev, section_id) {
            window.location.href = L.url('admin/services/dormnet/account/edit', section_id);
            return Promise.resolve();
        };

        return m.render();
    }
});
