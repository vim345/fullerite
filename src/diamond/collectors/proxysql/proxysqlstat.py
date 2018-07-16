# coding=utf-8
"""This code is a fork of
https://github.com/sysown/proxysql/blob/9dab0eba12a717b738264d6928599034ea3ebe81/diamond/proxysqlstat.py
"""
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
            'hosts': 'List of hosts to collect from. Format is yourusername:yourpassword@host:port/db',
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(ProxySQLCollector, self).get_default_config()
        config.update({
            'path': 'proxysql',
            'hosts': [],
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

    def _execute_mysql_status_query(self):
        return self.get_db_stats('SHOW MYSQL STATUS')

    def _execute_connection_pool_stats_query(self):
        rows = self.get_db_stats(
            'SELECT * FROM stats_mysql_connection_pool'
        )
        return rows

    def _is_number(self, number):
        try:
            float(number)
            return True
        except ValueError:
            return False

    def get_db_global_status(self):
        stats = []
        rows = self._execute_mysql_status_query()
        for row in rows:
            if self._is_number(row['Value']):
                stats.append(Metric(name=row['Variable_name'], value=float(row['Value']), dimensions=None))

        return stats

    def get_mysql_connection_pool_stats(self):
        rows = self._execute_connection_pool_stats_query()

        stats = []
        for row in rows:
            for metric_name, value in row.items():
                if self._is_number(value):
                    stats.append(
                        Metric(
                            name=metric_name,
                            value=value,
                            dimensions={
                                'hostgroup': row['hostgroup'],
                                'srv_host': row['srv_host'],
                            },
                        )
                    )

        return stats

    def get_stats(self, params):
        metrics = {}

        if not self.connect(params):
            return metrics

        metrics['status'] = self.get_db_global_status()
        metrics['stats_connection_pool'] = self.get_mysql_connection_pool_stats()

        self.disconnect()

        return metrics

    def _publish_stats(self, metrics):
       self._publish_status_metrics(metrics['status'])
       self._publish_connection_pool_metrics(metrics['stats_connection_pool'])

    def _publish_connection_pool_metrics(self, metrics):
        for metric in metrics:
            self._publish_proxysql_metric(metric)

    def _publish_status_metrics(self, metrics):
        for metric in metrics:
            if metric.name not in self.MYSQL_STATS_GLOBAL:
                metric = metric._replace(value=self.derivative(metric.name, metric.value))

            self._publish_proxysql_metric(metric)

    def _publish_proxysql_metric(self, metric):
        """Converts from our Metric datastructure into the call format of self.publish"""
        self.dimensions = metric.dimensions
        self.publish(metric.name, metric.value)

    def collect(self):
        if MySQLdb is None:
            self.log.error('Unable to import MySQLdb')
            return False

        for host in self.config['hosts']:
            try:
                metrics = self.get_stats(params=self.parse_host_config(host))
            except Exception as e:
                try:
                    self.disconnect()
                except MySQLdb.ProgrammingError:
                    pass
                self.log.error('Collection failed for {}'.format(e))
                continue

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
