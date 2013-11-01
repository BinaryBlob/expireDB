#!/usr/bin/env python3

import socket
import sys

for i in range(int(sys.argv[3])):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect((sys.argv[1], int(sys.argv[2])))

    sock.sendall('set {0} testbericht \n'.format(i))

    sock.sendall('get {0} \n'.format(i))
    res = sock.recv(4096)
    sock.close()
    print(res)


