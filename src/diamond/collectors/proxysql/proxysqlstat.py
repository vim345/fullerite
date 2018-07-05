# coding=utf-8
import diamond.collector
import re
import time
from collections import namedtuple

try:
    import MySQLdb
    from MySQLdb import MySQLError
except ImportError:
    MySQLdb = None
    MySQLError = ValueError


Metric = namedtuple('Metric', ['name', 'value', 'dimensions'])


class ProxySQLCollector(diamond.collector.Collector):

    MYSQL_STATS_GLOBAL = [
        'Active_Transactions',
        'Questions',

        'Client_Connections_aborted',
        'Client_Connecitons_created',
        'Client_Connections_connected',
        'Client_Connections_non_idle',

        'ConnPool_memory_bytes',
        'ConnPool_get_conn_immediate',
        'ConnPool_get_conn_success',
        'ConnPool_get_conn_failure',

        'MySQL_Monitor_Workers',
        'MySQL_Thread_Workers',

        'Backend_query_time_nsec',
        'Queries_backends_bytes_recv',
        'Queries_backends_bytes_sent',
        'Query_Processor_time_nsec',
        'Slow_queries'

        'SQLite3_memory_bytes',
        'ProxySQL_Uptime',

        'Server_Connections_aborted',
        'Server_Connections_connected',
        'Server_Connections_created',
        'Server_Connections_delayed',

        'mysql_backend_buffers_bytes',
        'mysql_frontend_buffers_bytes',
        'mysql_session_internal_bytes',
    ]

    def __init__(self, *args, **kwargs):
        super(ProxySQLCollector, self).__init__(*args, **kwargs)

    def process_config(self):
        super(ProxySQLCollector, self).process_config()
        if self.config['hosts'].__class__.__name__ != 'list':
            self.config['hosts'] = [self.config['hosts']]

    def get_default_config_help(self):
        config_help = super(ProxySQLCollector, self).get_default_config_help()
        config_help.update({
            'publish':
                "Which metrics you would like to publish. Leave unset to publish all",
            'hosts': 'List of hosts to collect from. Format is ' +
            'yourusername:yourpassword@host:port/db'
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(ProxySQLCollector, self).get_default_config()
        config.update({
            'path':     'proxysql',
            # Connection settings
            'hosts':    [],

        })
        return config

    def get_db_stats(self, query):
        cursor = self.db.cursor(cursorclass=MySQLdb.cursors.DictCursor)

        try:
            cursor.execute(query)
            return cursor.fetchall()
        except MySQLError, e:
            self.log.error('ProxySQLCollector could not get db stats', e)
            return ()

    def connect(self, params):
        try:
            self.db = MySQLdb.connect(**params)
            self.log.debug('ProxySQLCollector: Connected to database.')
        except MySQLError, e:
            self.log.error('ProxySQLCollector couldnt connect to database %s', e)
            return False
        return True

    def disconnect(self):
        self.db.close()

    def get_db_global_status(self):
        stats = {'status': {}}
        rows = self.get_db_stats('SHOW MYSQL STATUS')
        for row in rows:
            try:
                stats['status'][row['Variable_name']] = float(row['Value'])
            except:
                pass
        return stats

    def get_mysql_connection_pool_stats(self):
        rows = self.get_db_stats(
            'SELECT hostgroup, srv_host, ConnUsed, ConnFree, Latency_us FROM stats.stats_connection_pool'
        )
        for row in rows:
            try:
                stats['status'][row['Variable_name']] = float(row['Value'])
            except:
                pass
        return stats

    def get_stats(self, params):
        metrics = {}

        if not self.connect(params):
            return metrics

        global_stats = self.get_db_global_status()

        metrics.update(global_stats)

        self.disconnect()

        return metrics

    def _publish_stats(self, metrics):
       self._publish_status_metrics(metrics['status'])

    def _publish_status_metrics(self, metrics):
        for metric_name, metric_value in metrics.items():
            if type(metric_value) is not float:
                continue

            if metric_name not in self.MYSQL_STATS_GLOBAL:
                metric_value = self.derivative(metric_name, metric_value)
            if (('publish' not in self.config or
                 metric_name in self.config['publish'])):
                self.publish(metric_name, metric_value)

    def collect(self):
        if MySQLdb is None:
            self.log.error('Unable to import MySQLdb')
            return False

        for host in self.config['hosts']:
            try:
                metrics = self.get_stats(params=self.parse_host_config(host))
            except Exception, e:
                try:
                    self.disconnect()
                except MySQLdb.ProgrammingError:
                    pass
                self.log.error('Collection failed for %s %s', e)
                continue

            # Warn if publish contains an unknown variable
            if 'publish' in self.config and metrics['status']:
                for k in self.config['publish'].split():
                    if k not in metrics['status']:
                        self.log.error("No such key '%s' available, issue " +
                                       "'show global status' for a full " +
                                       "list", k)
            self._publish_stats(metrics)

    def parse_host_config(self, host):
        """Parses the host config to get the database connection string.

        Format is 'yourusername:yourpassword@host:port/db'"""
        matches = re.search('^([^:]*):([^@]*)@([^:]*):?([^/]*)/([^/]*)', host)

        if not matches:
            raise ValueError('Connection string {} is not in the required format'.format(host))

        params = {}

        params['user'] = matches.group(1)
        params['passwd'] = matches.group(2)
        params['host'] = matches.group(3)
        try:
            params['port'] = int(matches.group(4))
        except ValueError:
            params['port'] = 3306
        params['db'] = matches.group(5)
        return params
