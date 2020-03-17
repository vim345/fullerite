import json
import urllib2

KUBELET_DEFAULT_PORT = 10255
KUBELET_DEFAULT_HOST = "localhost"


class Kubelet(object):

    def list_pods(self):
        url = "http://{}:{}/pods".format(KUBELET_DEFAULT_HOST, KUBELET_DEFAULT_PORT)
        try:
            response = urllib2.urlopen(url)
        except urllib2.HTTPError as err:
            return None, err

        try:
            return json.load(response), None
        except (TypeError, ValueError) as e:
            return None, e
