import http.server
import ssl

print("server running at: 4443")

httpd = http.server.HTTPServer(
    ('localhost', 4443), http.server.SimpleHTTPRequestHandler)
httpd.socket = ssl.wrap_socket(
    httpd.socket, certfile='./cert.pem', keyfile='./key.pem', server_side=True)
httpd.serve_forever()
