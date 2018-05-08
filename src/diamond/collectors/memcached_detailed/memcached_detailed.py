# coding=utf-8

"""
Collect memcached detailed stats

Detailed memcache stats can only be collected if memcached detailed stats
are turned on.

#### Dependencies

 * subprocess

#### Example Configuration

MemcachedDetailedCollector.conf

```
    enabled = True
    hosts = localhost:11211, app-1@localhost:11212, app-2@localhost:11213, etc
```

TO use a unix socket, set a host string like this

```
    hosts = /path/to/blah.sock, app-1@/path/to/bleh.sock,
```
"""

import diamond.collector
import socket
import re

from collections import namedtuple

PREFIX_PATTERN = 'PREFIX (?P<cache_name>[\w\-\.]+)'
STATS_PATTERN = ' get (?P<get>[0-9]+) hit (?P<hit>[0-9]+) set (?P<set>[0-9]+) del (?P<del>[0-9]+)$'

HOST_REGEX = re.compile('((?P<alias>.+)\@)?(?P<hostname>[^:]+)(:(?P<port>\d+))?')
LINE_REGEX = re.compile(PREFIX_PATTERN + STATS_PATTERN)

Stat = namedtuple('Stat', ['cache_name', 'detailed_get', 'detailed_hit', 'detailed_set', 'detailed_del'])
stat_fields = set(['detailed_get', 'detailed_hit', 'detailed_set', 'detailed_del'])

class MemcachedDetailedCollector(diamond.collector.Collector):

    def get_default_config_help(self):
        config_help = super(MemcachedDetailedCollector, self).get_default_config_help()
        config_help.update({
            'hosts': "List of hosts, and ports to collect. Set an alias by "
            + " prefixing the host:port with alias@",
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(MemcachedDetailedCollector, self).get_default_config()
        config.update({
            'path': 'memcached_detailed',
            # Connection settings
            'hosts': ['localhost:16666']
        })
        return config

    def get_raw_stats(self, host, port):
        data = ''
        # connect
        try:
            if port is None:
                sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
                sock.connect(host)
            else:
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.connect((host, int(port)))
            # request stats, if detailed stats not turned on data will be empty
            sock.send('stats detail dump\n')
            # something big enough to get whatever is sent back
            data = sock.recv(4096)
        except socket.error:
            self.log.error('Failed to get detailed stats from %s:%s',
                               host, port)
        return data

    def get_stats(self, host, port):
        stats = set()

        data = self.get_raw_stats(host, port)

        for line in data.splitlines():
            match = LINE_REGEX.match(line)
            if match:
                stats.add(Stat(
                    cache_name=match.group('cache_name'),
                    detailed_get=int(match.group('get')),
                    detailed_hit=int(match.group('hit')),
                    detailed_set=int(match.group('set')),
                    detailed_del=int(match.group('del')),
                ))
        return stats

    def collect(self):
        hosts = self.config.get('hosts')

        # Convert a string config value to be an array
        if isinstance(hosts, basestring):
            hosts = [hosts]

        for host in hosts:
            match = HOST_REGEX.match(host)
            alias = match.group('alias')
            hostname = match.group('hostname')
            port = match.group('port')

            if alias is None:
                alias = hostname

            stats = self.get_stats(hostname, port)

            for stat in stats:
                for field in stat_fields:
                    self.dimensions = {
                        'memcache_host': alias,
                        'cache_name': stat.cache_name,
                    }
                    metric_name = '.'.join(['memcache', field])
                    self.publish_cumulative_counter(metric_name, getattr(stat, field))
