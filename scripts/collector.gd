extends Node

signal connected
signal disconnected
signal error

signal event

var _status: int = 0
var _stream: StreamPeerTCP = StreamPeerTCP.new()

var host = "127.0.0.1"
var port = 1337

var _sensor_pid = -1

# Called when the node enters the scene tree for the first time.
func _ready():
	_status = _stream.get_status()
	_sensor_pid = OS.create_process("./bin/sensor", [])
	OS.delay_msec(1000)
	connect_to_sensor()

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	_stream.poll()
	var new_state = _stream.get_status()
	if new_state != _status:
		# something has changed
		_status = new_state
		match _status:
			_stream.STATUS_NONE:
				print("disconnected :(")
				emit_signal("disconnected")
			_stream.STATUS_CONNECTING:
				print("connecting!")
			_stream.STATUS_CONNECTED:
				print("connected!")
				emit_signal("connected")
			_stream.STATUS_ERROR:
				print("failed for some reason")
				emit_signal("error")
	var available = _stream.get_available_bytes()
	if available > 0:
		var proc = _stream.get_partial_data(available)
		if proc[0] != OK:
			print("failed to read " + str(proc[0]))
			emit_signal("error")
		else:
			for b in proc[1]:
				match b:
					0x01:
						emit_signal("event", "process")
					0x02:
						emit_signal("event", "network")
					0x03:
						emit_signal("event", "endpoint")
				
func connect_to_sensor():
	_status = _stream.STATUS_NONE
	if _stream.connect_to_host(host, port) != OK:
		emit_signal("error")

func _exit_tree():
	if _sensor_pid != -1:
		#OS.kill(_sensor_pid)
		pass
