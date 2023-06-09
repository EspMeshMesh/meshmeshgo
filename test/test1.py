import socket
import binascii
import time

# Connect and send an hello

# create a socket
client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client.settimeout(5)
# connect to the server
client.connect(("127.0.0.1", 6053))

client.send(b"INIT|0.28.139.56|6053\n")

# receive
response = client.recv(4096)
print(response)

if response == b"!!OK!":
    time.sleep(0.5)
    client.send(binascii.unhexlify(b"0013010a0d61696f657370686f6d6561706910011807"))

    while True:
        try:
            response = client.recv(4096)
            if response == b"\x00\x00\x07":
                client.send(b"\x00\x00\x08")
            print(binascii.hexlify(response))
        except KeyboardInterrupt:
            break
        except TimeoutError:
            time.sleep(0.1)

while True:
    time.sleep(1)

client.close()
