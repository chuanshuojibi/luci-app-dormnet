#!/usr/bin/env ucode

'use strict';

import { popen } from 'fs';

function json_process(command) {
    let profile = {
        "success": false,
        "message": "failed to open process in ucode",
    };
    const process = popen(command);
    if (process) {
        profile = json(process);
        process.close();
    }
    return profile;
}

function iface_device(iface) {
    if (!match(iface, /^[A-Za-z0-9_.:-]+$/)) {
        return "";
    }

    const command = `ubus call network.interface.${iface} status 2>/dev/null`;
    const process = popen(command);
    if (!process) {
        return iface;
    }

    const status = json(process);
    process.close();

    return status?.l3_device || status?.device || iface;
}

function ping_iface(iface, target) {
    const device = iface_device(iface);
    if (!device || !match(device, /^[A-Za-z0-9_.:-]+$/)) {
        return {
            "success": false,
            "message": "invalid interface name",
        };
    }

    const command = `ping -I ${device} -c 3 -W 2 ${target} >/dev/null 2>&1 && echo '{"success":true,"message":"connected"}' || echo '{"success":false,"message":"unreachable"}'`;
    return json_process(command);
}

const methods = {
    build_info: {
        call: function() {
            const command = `dormnet -build-info`;
            return json_process(command);
        }
    },

    extra_args_account: {
        args: { account: 'account' },
        call: function(req) {
            const account = req.args?.account ?? "";
            const command = `dormnet -extra-args-account=${account}`;
            return json_process(command);
        }
    },

    list_target: {
        call: function() {
            const command = `dormnet -list-targets`;
            return json_process(command);
        }
    },

    status: {
        call: function() {
            const command = `dormnet -status`;
            return json_process(command);
        }
    },

    logs: {
        call: function() {
            const command = `dormnet -logs`;
            return json_process(command);
        }
    },

    ping_internet: {
        args: { iface: 'iface' },
        call: function(req) {
            const iface = req.args?.iface ?? "";
            return ping_iface(iface, "8.8.8.8");
        }
    },

    ping_campus: {
        args: { iface: 'iface' },
        call: function(req) {
            const iface = req.args?.iface ?? "";
            return ping_iface(iface, "192.168.200.2");
        }
    },

    restart: {
        call: function () {
            const command = `/etc/init.d/dormnet restart`;
            return json_process(command);
        }
    }
};

return { 'luci.dormnet': methods };
