# coding=utf-8

"""
Port of the Ganglia CPU Model collector.
Collects the no. of physical cores and the model name from /proc/cpuinfo
"""

import diamond.collector


class CPUModelCollector(diamond.collector.Collector):

    PROC_CPU_INFO = "/proc/cpuinfo"

    def get_default_config_help(sef):
        config_help = super(CPUModelCollector, self).get_default_config_help()
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(CPUModelCollector, self).get_default_config()
        return config

    def collect(self):
        """
        Collector cpu model
        """
        def get_processor_model():
            phys_ids = {}
            model_name = None
            f = None
            try:
                f = file(self.PROC_CPU_INFO)
                for line in f:
                    if line.startswith('physical id\t'):
                        phys_id = int(line.split(':')[1].strip())
                        phys_ids[phys_id] = None
                    elif line.startswith('model name\t'):
                        if model_name is None:
                            model_name = line.split(':')[1].strip()
                        else:
                            if model_name != model_name:
                                model_name = 'mixed'

                # need to shorten model name, we can only return 32 chars
                rs = []
                for s in model_name.split():
                    if s not in ('AMD', 'Processor', 'Intel(R)', 'CPU'):
                        for suf in ('(tm)', '(TM)'):
                            if s.endswith(suf):
                                s = s.rstrip(suf)
                        rs.append(s)
                model_name = ' '.join(rs)

                self.log.debug('%dx %s', len(phys_ids), model_name)
                return (len(phys_ids), model_name)
            finally:
                f.close()

        try:
            # Collect & Publish Metrics
            stats = get_processor_model()
            metric_name = 'cpu_model.{0}'.format(stats[1])
            self.publish(metric_name, stats[0])
            return True
        except Exception, e:
            self.log("Error: %s", e)
            return False
