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

function shell_read(command) {
    const p = popen(command);
    if (!p) return "";
    let buf = "";
    let chunk;
    while ((chunk = p.read(4096)) !== null && chunk !== "") {
        buf += chunk;
    }
    p.close();
    return buf;
}

function iface_device(iface) {
    if (!match(iface, /^[A-Za-z0-9_.:-]+$/)) {
        return "";
    }

    // 1) 当成 network 名字查 ubus
    const out = shell_read(`ubus call network.interface.${iface} status 2>/dev/null`);
    if (out) {
        try {
            const status = json(out);
            const dev = status?.l3_device || status?.device;
            if (dev) return dev;
        } catch (e) {}
    }

    // 2) 退而求其次：直接当设备名用（用户填了 eth1/br-lan 这种）
    const exists = shell_read(`[ -d /sys/class/net/${iface} ] && echo yes`);
    if (trim(exists) === "yes") return iface;

    // 3) 找不到设备
    return "";
}

function ping_iface(iface, target) {
    const device = iface_device(iface);
    if (!device || !match(device, /^[A-Za-z0-9_.:-]+$/)) {
        return {
            "success": false,
            "message": `interface "${iface}" has no resolvable L3 device`,
            "data": {
                "iface": iface,
                "device": "",
                "target": target,
                "output": "",
                "loss": 100,
            },
        };
    }

    // 把 stderr 也合到 stdout，原样返回给 UI 展示
    const cmd = `ping -I ${device} -c 4 -W 2 ${target} 2>&1`;
    const output = shell_read(cmd);

    // 解析丢包率
    let loss = 100;
    const m = match(output, /([0-9]+)% packet loss/);
    if (m) loss = +m[1];
    const success = loss < 100;

    return {
        "success": success,
        "message": success ? `reachable (${100 - loss}% replied)` : "unreachable",
        "data": {
            "iface": iface,
            "device": device,
            "target": target,
            "output": output,
            "loss": loss,
        },
    };
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

    account_status: {
        call: function() {
            const command = `dormnet -account-status`;
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
