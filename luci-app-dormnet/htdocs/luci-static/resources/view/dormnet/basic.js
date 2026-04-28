'use strict';
'require view';
'require form';
'require poll';
'require uci';
'require tools.dormnet as dormnet';

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
        ]);
    },
    render: function(data) {
        const status = data[0];
        const buildInfo = data[1];

        let m, s, o;

        m = new form.Map('dormnet', `${_('DormNet')} v${buildInfo.data.version}`,
            `${_('Manage your campus network with DormNet on OpenWrt.')}`);

        s = m.section(form.NamedSection, 'basic', 'basic');

        o = s.option(form.Flag, 'enabled', _('Enabled'));
        o.default = false;

        o = s.option(form.Button, 'restart', _('Restart'));
        o.description = _("Restart manually")
        o.depends('enabled', '1');

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

        o = s.option(form.ListValue, 'login_account', _('Login account'));
        o.description = _('Select the saved account used for campus network login.');
        o.optional = true;
        o.value('', _('All accounts'));
        for (const account of uci.sections('dormnet', 'account')) {
            o.value(account['.name'], account.username || account['.name']);
        }

        o = s.option(form.Flag, 'use_sd_network', _('Use software-defined networking'));
        o.description = _("Select when you need multiple geographically dispersed devices to use the campus network simultaneously.")
        o.default = '0';

        o = s.option(form.ListValue, 'work_with', _('Work with'));
        o.default = 'master';
        o.depends('use_sd_network', '1')
        o.value('easytier', _('EasyTier'));
        // o.value('zerotier', _('ZeroTier'));

        o = s.option(form.Value, 'listen', _('Listen address'));
        o.depends({
            'work_mode': 'master',
            'work_with': 'easytier'
        });
        o.default = "127.0.0.1:10721";

        return m.render();
    }
});
