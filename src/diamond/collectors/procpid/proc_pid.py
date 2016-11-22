# coding=utf-8

"""
Collect /proc/{pid}/ stats for a specific set of pids


"""
import diamond.collector
import os
import subprocess


class ProcPidCollector(diamond.collector.Collector):

    def get_default_config_help(self):
        config_help = super(ProcPidCollector, self).get_default_config_help()
        config_help.update({
            'pid_paths': 'Dict of name to pid file',
            'ls': 'Path to ls command',
            'sudo_cmd': 'Path to sudo',
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(ProcPidCollector, self).get_default_config()
        config.update({
            'pid_paths': {},
            'ls': '/bin/ls',
            'sudo_cmd': '/usr/bin/sudo',
        })
        return config

    def get_proc_path(self, pid_path):
        with open(pid_path) as fp:
            pid = fp.read().strip()
        proc_path = os.path.join('/proc', pid)
        return proc_path

    def get_fds(self, proc_path):
        proc_path = os.path.join(proc_path, 'fd')

        command = [self.config['ls'], proc_path]
        command.insert(0, self.config['sudo_cmd'])

        process = subprocess.Popen(command,
                                   stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE)
        output, errors = process.communicate()
        if errors:
            raise Exception(errors)
        return len(output.splitlines())

    def collect(self):
        """
        Collector pid_proc stats
        """
        for service, pid_path in self.config['pid_paths'].items():
            fds = self.get_fds(self.get_proc_path(pid_path))
            try:
                self.dimensions = {'service': service}
                self.publish('proc_pid_stats.fds', fds)
            except Exception, e:
                self.log.error("ProcPidCollector Error: %s", e)
