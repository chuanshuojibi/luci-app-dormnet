'use strict';
'require baseclass';
'require uci';
'require fs';
'require rpc';
'require request';

const callDormnetRestart = rpc.declare({
    object: 'luci.dormnet',
    method: 'restart',
    expect: { '': {} }
});

const callDormnetBuildInfo = rpc.declare({
    object: 'luci.dormnet',
    method: 'build_info',
    expect: { '': {} }
});

const callDormnetStatus = rpc.declare({
    object: 'luci.dormnet',
    method: 'status',
    expect: { '': {} }
});

const callDormnetSupportedTargets = rpc.declare({
    object: 'luci.dormnet',
    method: 'list_target',
    expect: { '': {} }
});

const callDormnetExtraArgsAccount = rpc.declare({
    object: 'luci.dormnet',
    method: 'extra_args_account',
    params: [ 'account' ],
    expect: { '': {} }
});

const callDormnetLogs = rpc.declare({
    object: 'luci.dormnet',
    method: 'logs',
    expect: { '': {} }
});

const callDormnetAccountStatus = rpc.declare({
    object: 'luci.dormnet',
    method: 'account_status',
    expect: { '': {} }
});

const callDormnetPingInternet = rpc.declare({
    object: 'luci.dormnet',
    method: 'ping_internet',
    params: [ 'iface' ],
    expect: { '': {} }
});

const callDormnetPingCampus = rpc.declare({
    object: 'luci.dormnet',
    method: 'ping_campus',
    params: [ 'iface' ],
    expect: { '': {} }
});

// noinspection JSAnnotator
return baseclass.extend({
    buildInfo: function () {
        return callDormnetBuildInfo();
    },
    status: function () {
        return callDormnetStatus();
    },
    supportedTargets: function () {
        return callDormnetSupportedTargets();
    },
    extraArgsAccount: function (account) {
        return callDormnetExtraArgsAccount(account);
    },
    logs: function () {
        return callDormnetLogs();
    },
    accountStatus: function () {
        return callDormnetAccountStatus();
    },
    pingInternet: function (iface) {
        return callDormnetPingInternet(iface);
    },
    pingCampus: function (iface) {
        return callDormnetPingCampus(iface);
    },
    restart: function () {
        return callDormnetRestart();
    }
});
