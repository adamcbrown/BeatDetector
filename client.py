from threading import Thread
from time import sleep
import socket

import subprocess
proc = subprocess.Popen(['go','run','main.go'],stdout=subprocess.PIPE)

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

while True:
	try:
		sock.connect(("127.0.0.1", 8000))
		while True:
			print("NOW")
			sock.send("\n".encode())
			print(proc.stdout.readline())
	except socket.error as exc:
		sock.close()
		sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
		pass
	except KeyboardInterrupt:
		sock.send("x".encode())
		sock.close()
		break
