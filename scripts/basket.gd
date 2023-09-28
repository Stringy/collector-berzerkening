extends StaticBody2D

signal captured_process
signal captured_network
signal captured_endpoint

@export var speed: float = 500.0

# Called when the node enters the scene tree for the first time.
func _ready():
	pass # Replace with function body.

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	var vec = get_viewport().get_mouse_position()
	vec.y = self.position.y
	position = vec

func _on_area_2d_body_shape_entered(body_rid, body, body_shape_index, local_shape_index):
	if body.is_in_group("processes"):
		emit_signal("captured_process")
	elif body.is_in_group("networks"):
		emit_signal("captured_network")
	elif body.is_in_group("endpoints"):
		emit_signal("captured_endpoint")
	body.queue_free()
