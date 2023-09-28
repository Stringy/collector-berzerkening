extends Control


# Called when the node enters the scene tree for the first time.
func _ready():
	if OS.get_name() == "HTML5":
		DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	pass

func _input(event):
	if event.is_action_pressed("ui_cancel") and OS.get_name() == "HTML5":
		get_tree().quit()
