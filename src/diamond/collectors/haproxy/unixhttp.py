# coding=utf-8

"""
urllib2 unix HTTP handler
"""

import urllib2
import httplib
import socket

class UnixHTTPResponse(httplib.HTTPResponse, object):

    def __init__(self, sock, *args, **kwargs):
        disable_buffering = kwargs.pop('disable_buffering', False)
        kwargs['buffering'] = not disable_buffering
        super(UnixHTTPResponse, self).__init__(sock, *args, **kwargs)


class UnixHTTPConnection(httplib.HTTPConnection, object):

    def __init__(self, unix_socket, timeout=socket._GLOBAL_DEFAULT_TIMEOUT):
        super(UnixHTTPConnection, self).__init__(
            'localhost', timeout=timeout
        )
        self.unix_socket = unix_socket
        self.timeout = timeout
        self.disable_buffering = False

    def connect(self):
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        if self.timeout is not socket._GLOBAL_DEFAULT_TIMEOUT:
            sock.settimeout(self.timeout)
        sock.connect(self.unix_socket)
        self.sock = sock

    def putheader(self, header, *values):
        super(UnixHTTPConnection, self).putheader(header, *values)
        if header == 'Connection' and 'Upgrade' in values:
            self.disable_buffering = True

    def response_class(self, sock, *args, **kwargs):
        if self.disable_buffering:
            kwargs['disable_buffering'] = True

        return UnixHTTPResponse(sock, *args, **kwargs)


class UnixHTTPHandler(urllib2.AbstractHTTPHandler):

    def unixhttp_open(self, req):
        return self.do_open(UnixHTTPConnection, req)

    unixhttp_request = urllib2.AbstractHTTPHandler.do_request_
