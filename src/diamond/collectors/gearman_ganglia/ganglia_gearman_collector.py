# coding=utf-8

"""
Port of the ganglia gearman collector
"""

import diamond.collector
import os
import subprocess
import time

import gearman


class GearmanCollector(diamond.collector.Collector):

    GEARMAN_PID_PATH_DEFAULT="/var/run/gearman/gearman-job-server.pid"
    GEARMAN_ENDPOINT_DEFAULT="localhost"

    def get_default_config_help(sef):
        config_help = super(GearmanCollector, self).get_default_config_help()
        config_help.update({
            'gearman_pid_path': 'Gearman PID file path',
            'url': 'Gearman endpoint to talk to',
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(GearmanCollector, self).get_default_config()
        config.update({
            'gearman_pid_path': self.GEARMAN_PID_PATH_DEFAULT,
            'url': self.GEARMAN_ENDPOINT_DEFAULT,
        })
        return config

    def collect(self):
        """
        Collector gearman stats
        """
        def gearman_ping(gm_admin_client):
            server_ping = gm_admin_client.ping_server()
            return server_ping

        def gearman_queued(gm_admin_client):
            server_status = gm_admin_client.get_status()
            queued = 0
            for entry in server_status:
                queued += entry['queued']
            return queued

        def get_fds(gearman_pid_path):
            with open(gearman_pid_path) as fp:
                gearman_pid = fp.read().strip()
            gearman_dir = os.path.join("/proc/", gearman_pid, "fd/")
            
            process = subprocess.Popen('sudo ls ' + gearman_dir,
                          shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            count = len(process.stdout.readlines())
            return count

        try:
            # Collect and Publish Metrics
            self.log.debug("Using pid file: %s & gearman endpoint : %s",
                    self.config['gearman_pid_path'], self.config['url'])
            
            gm_admin_client = gearman.GearmanAdminClient([self.config['url']])
            self.publish('gearman_ping', gearman_ping(gm_admin_client))
            self.publish('gearman_queued', gearman_queued(gm_admin_client))
            self.publish('gearman_fds', get_fds(self.config['gearman_pid_path']))
        except Exception, e:
            self.log.error("Error: %s", e)
