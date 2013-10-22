#!/usr/bin/env python3

import socket
import sys

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect((sys.argv[1], sys.argv[2]))

for i in range(10):
    sock.sendall('set {0} testbericht \n'.format(i))

    sock.sendall('get {0} \n'.format(i))
    res = sock.recv(4096)
    #print(res)

sock.close()
