from threading import Thread
from time import sleep
import socket

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.bind(('127.0.0.1', 8000))
sock.listen(1)
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

h = 0
s = 1
v = 1

def update_lights():
	global h
	global s
	global v
	update_rate = 0.02

	while True:
		if v > 0.5:
			print(h, s, v)
			v -= 0.5 * update_rate
			if v < 0.5:
				v = 0.5

			# Call light set function here
			sleep(update_rate)

thread = Thread(target = update_lights)
thread.start()

while True:
	conn = None
	try:
		print("Accepting...")
		conn, addr = sock.accept()
		print("Accepted")
		while True:
			conn.settimeout(5)
			char = conn.recv(1).decode()
			if char == '\n':
				h = (h+120)%360
				v = 1
			elif char == "x":
				conn.close()
				break;
	except socket.error as exc:
		if conn != None:
			conn.close()