'use strict';
'require view';
'require ui';
'require poll';
'require tools.dormnet as dormnet';

function toText(resp) {
    const arr = resp && Array.isArray(resp.data) ? resp.data : [];
    return arr.length > 0 ? arr.join('\n') : _('No logs yet.');
}

// noinspection JSAnnotator
return view.extend({
    load: function () {
        return Promise.all([
            dormnet.logs(),
        ]);
    },
    render: function(data) {
        const initialText = toText(data[0]);

        const ta = new ui.Textarea(initialText, {
            readonly: true,
            monospace: true,
            rows: 30,
        });

        const taNode = ta.render();

        function autoScroll() {
            const inner = taNode.querySelector('textarea');
            if (inner) inner.scrollTop = inner.scrollHeight;
        }
        // 首次滚到底
        setTimeout(autoScroll, 0);

        poll.add(function () {
            return L.resolveDefault(dormnet.logs()).then(function (resp) {
                ta.setValue(toText(resp));
                autoScroll();
            });
        });

        const refreshBtn = E('button', {
            'class': 'btn cbi-button cbi-button-action',
            'style': 'margin-bottom:.5em;',
            'click': function (ev) {
                ev.preventDefault();
                return L.resolveDefault(dormnet.logs()).then(function (resp) {
                    ta.setValue(toText(resp));
                    autoScroll();
                });
            }
        }, _('Refresh'));

        const clearBtn = E('button', {
            'class': 'btn cbi-button cbi-button-remove',
            'style': 'margin-bottom:.5em;margin-left:.5em;',
            'click': function (ev) {
                ev.preventDefault();
                ta.setValue('');
            }
        }, _('Clear view'));

        return E('div', {}, [
            E('h2', {}, _('DormNet Logs')),
            E('div', { 'class': 'cbi-section-descr' },
                _('Recent dormnet syslog entries (auto-refreshes every few seconds).')),
            E('div', {}, [refreshBtn, clearBtn]),
            taNode,
        ]);
    }
});
