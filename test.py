#!/usr/bin/env python3

import socket

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(('127.0.0.1', 11211))

#for i in range(100000):
#    sock.sendall('set {0} testbericht \n'.format(i))

for i in range(100000):
    sock.sendall('get {0} \n'.format(i))
    res = sock.recv(4096)
    #print(res)

sock.close()
