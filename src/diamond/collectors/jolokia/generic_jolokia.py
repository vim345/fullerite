# coding=utf-8

"""
Jolokia generic collector collects JMX metrics from jolokia agent with support
for dimensions
"""
from jolokia import JolokiaCollector

class GenericJolokiaCollector(JolokiaCollector):
    def collect_bean(self, prefix, obj):
        for k, v in obj.iteritems():
            if type(v) in [int, float, long]:
                self.parse_dimension_bean(prefix, k, v)
            elif isinstance(v, dict):
                self.collect_bean("%s.%s" % (prefix, k), v)
            elif isinstance(v, list):
                self.interpret_bean_with_list("%s.%s" % (prefix, k), v)

    def patch_dimensions(self, bean, dims):
        metric_name = dims.pop("name", None)
        return metric_name, None, dims

    def patch_metric_name(self, bean, metric_name_list):
        if self.config.get('prefix', None):
            metric_name_list = [self.config['prefix']] + metric_name_list

        lower_bean_key = bean.bean_key.lower()

        if lower_bean_key not in ['count', 'value']:
            metric_name_list.append(lower_bean_key)

        return metric_name_list
