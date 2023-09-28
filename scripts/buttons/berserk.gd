extends Button

@export var demo: bool = false

# Called when the node enters the scene tree for the first time.
func _ready():
	pass

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	pass


func _on_pressed():
	if not demo:
		get_tree().change_scene_to_file("res://scenes/hard-mode.tscn")
	else:
		get_tree().change_scene_to_file("res://scenes/berserker.tscn")
