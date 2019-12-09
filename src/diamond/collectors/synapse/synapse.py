"""
Collect the number of haproxy frontends and backends for synapse.
"""
import diamond.collector
import json

class SynapseCollector(diamond.collector.Collector):
    def get_default_config_help(self):
        config_help = super(SynapseCollector, self).get_default_config_help()
        config_help.update({
            'path': 'Path to synapse.conf.json',
        })
        return config_help


    def get_default_config(self):
        config = super(SynapseCollector, self).get_default_config()
        config.update({
            'path': '/etc/synapse/synapse.conf.json',
        })
        return config


    def collect(self):
        path = self.config.get('path')

        try:
            f = open(path)
            synapse_conf = json.loads(f.read())
        except ValueError:
            return
        finally:
            f.close()

        frontend_count = 0
        backend_count = 0

        for config in synapse_conf.get('services', {}).values():
            haproxy = config.get("haproxy", {})
            if haproxy.get("disabled"):
                continue
            if haproxy.get("backend"):
                backend_count += 1
            if haproxy.get("frontend"):
                frontend_count += 1

        self.publish("synapse.frontend_count", frontend_count)
        self.publish("synapse.backend_count", backend_count)
