extends Node2D

@export var simulate: bool = true

var ball = preload("res://scenes/projectile.tscn")
var explosion = preload("res://scenes/explosion.tscn")

var process_score = 0
var network_score = 0 

var per_second = 1
var max_per_second = 120

var queue = 0
var max_queue = 5

var timer: Timer

# Called when the node enters the scene tree for the first time.
func _ready():
	timer = Timer.new()
	timer.wait_time = 1.0
	add_child(timer)
	
	timer.connect("timeout", _on_timer_timeout)
	timer.start()
	
func _on_timer_timeout():
	max_per_second += 1

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	if simulate:
		_do_simulation()

func _do_simulation():
	if queue < max_queue:
		var instance = ball.instantiate()
		var x = randf_range(0, ProjectSettings.get_setting("display/window/size/viewport_width"))
		instance.position = Vector2(x, 0)
		
		if randi_range(1, 2) == 1:
			instance.set_event_type("process")
			instance.add_to_group("processes")
		else:
			instance.set_event_type("connection")
			instance.add_to_group("networks")
		
		queue += 1
		
		await get_tree().create_timer(1).timeout
		
		instance.connect("exploded", _on_projectile_exploded)
		add_child(instance)
		queue -= 1

func _input(event):
	if event.is_action_pressed("ui_cancel"):
		get_tree().change_scene_to_file("res://scenes/main.tscn")

func _on_basket_captured_process():
	$Scoreboard.process()
	
func _on_basket_captured_network():
	$Scoreboard.connection()
	
func _on_basked_captured_endpoint():
	$Scoreboard.endpoint()
	
func _on_collector_event(kind):
	var instance = ball.instantiate()
	var x = randf_range(0, ProjectSettings.get_setting("display/window/size/viewport_width"))
	instance.position = Vector2(x, 0)
	
	instance.set_event_type(kind)
	match kind:
		"process":
			instance.add_to_group("processes")
		"network":
			instance.add_to_group("network")
		"endpoint":
			instance.add_to_group("endpoints")

	instance.connect("exploded", _on_projectile_exploded)
	add_child(instance)

func _on_projectile_exploded(child):
	var instance = explosion.instantiate()
	instance.position = child.position - Vector2(0, 20)
	add_child(instance)
	child.queue_free()
