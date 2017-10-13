# coding=utf-8

"""
Collect zookeeper stats. ( Modified from memcached collector )

#### Dependencies

 * subprocess
 * Zookeeper 'mntr' command (zookeeper version => 3.4.0)

#### Example Configuration

ZookeeperCollector.conf

```
    enabled = True
    hosts = localhost:2181, app-1@localhost:2181, app-2@localhost:2181, etc
```

TO use a unix socket, set a host string like this

```
    hosts = /path/to/blah.sock, app-1@/path/to/bleh.sock,
```
"""

import diamond.collector
import socket
import re


class ZookeeperCollector(diamond.collector.Collector):

    def get_default_config_help(self):
        config_help = super(ZookeeperCollector, self).get_default_config_help()
        config_help.update({
            'publish': "Which rows of 'status' you would like to publish."
            + " Telnet host port' and type stats and hit enter to see the list"
            + " of possibilities. Leave unset to publish all.",
            'hosts': "List of hosts, and ports to collect. Set an alias by "
            + " prefixing the host:port with alias@",
            'reset_stats': "Reset the server stats via 'srst' command after"
            + "each collection. Enable this to avoid the max_latency statistic from"
            + "sticking to the high watermark since the process started."
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(ZookeeperCollector, self).get_default_config()
        config.update({
            'path':     'zookeeper',

            # Which rows of 'status' you would like to publish.
            # 'telnet host port' and type mntr and hit enter to see the list of
            # possibilities.
            # Leave unset to publish all
            # 'publish': ''

            # Connection settings
            'hosts': ['localhost:2181'],
            # reset the zk server stats after each collection?
            'reset_stats': False,
        })
        return config

    def get_raw_stats(self, host, port):
        """ Returns the raw zk mntr output """
        return self._zk_request('mntr', host, port)

    def _reset_zk_stats(self, host, port):
        """ Resets the zookeeper latency and byte stats """
        try:
            return self._zk_request('srst', host, port)
        except:
            self.log.exception("Caught exception resetting zk stats")

    def _zk_request(self, command, host, port):
        """ Performs the given zk four letter command and returns the output """
        data = ''
        # connect
        try:
            if port is None:
                sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
                sock.connect(host)
            else:
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.connect((host, int(port)))
            # request stats
            sock.send('%s\n' % command)
            # something big enough to get whatever is sent back
            data = sock.recv(4096)
        except socket.error:
            self.log.exception('Failed to send %s command to %s:%s',
                               command, host, port)
        return data

    def get_stats(self, host, port):
        # stuff that's always ignored, aren't 'stats'
        ignored = ('zk_version', 'zk_server_state')

        stats = {}
        data = self.get_raw_stats(host, port)

        # parse stats
        for line in data.splitlines():
            pieces = line.split()
            if pieces[0] in ignored:
                continue
            stats[pieces[0]] = pieces[1]

        return stats

    def collect(self):
        hosts = self.config.get('hosts')
        reset_stats = self.config.get('reset_stats')

        # Convert a string config value to be an array
        if isinstance(hosts, basestring):
            hosts = [hosts]

        for host in hosts:
            matches = re.search('((.+)\@)?([^:]+)(:(\d+))?', host)
            alias = matches.group(2)
            hostname = matches.group(3)
            port = matches.group(5)

            if alias is None:
                alias = hostname

            stats = self.get_stats(hostname, port)

            # figure out what we're configured to get, defaulting to everything
            desired = self.config.get('publish', stats.keys())

            # for everything we want
            for stat in desired:
                if stat in stats:
                    self.dimensions = {
                        'zookeeper_host': alias,
                    }
                    self.publish(stat, stats[stat])
                else:
                    # we don't, must be something configured in publish so we
                    # should log an error about it
                    self.log.error("No such key '%s' available, issue 'stats' "
                                   "for a full list", stat)

            # reset the stats so we get an average and max latency _since the last collection_
            if reset_stats:
                self._reset_zk_stats(hostname, port)
