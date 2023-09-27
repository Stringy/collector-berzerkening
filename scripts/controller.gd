extends Node2D

enum Mode {
	Game,
	Berserker,
	Collector
}

@export var mode: Mode = Mode.Game
@export var start_interval: float = 1
@export var min_interval: float = 0.1
@export var per_frame: float = 60

var ball = preload("res://scenes/projectile.tscn")
var explosion = preload("res://scenes/explosion.tscn")

var process_score = 0
var network_score = 0 

var per_second = 1
var max_per_second = 120

var queue = 0
var max_queue = 5

var spawn_timer: Timer
var difficulty_timer: Timer

var current_interval = 1.0

# Called when the node enters the scene tree for the first time.
func _ready():
	spawn_timer = Timer.new()
	spawn_timer.wait_time = start_interval
	
	difficulty_timer = Timer.new()
	difficulty_timer.wait_time = 5.0
	
	add_child(spawn_timer)
	add_child(difficulty_timer)
	
	difficulty_timer.connect("timeout", _on_difficulty_timeout)
	difficulty_timer.start()
	
	spawn_timer.connect("timeout", _on_timer_timeout)
	spawn_timer.start()
	
func _on_timer_timeout():
	if mode == Mode.Game:
		_do_simulation()
	
func _on_difficulty_timeout():
	if current_interval < min_interval:
		return
	current_interval /= 2
	spawn_timer.wait_time = current_interval 

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	if mode == Mode.Berserker:
		for i in range(per_frame):
			_do_simulation()
	
func _do_simulation():
	var instance = ball.instantiate()
	var x = randf_range(0, ProjectSettings.get_setting("display/window/size/viewport_width"))
	instance.position = Vector2(x, 0)
	
	match randi_range(0, 10):
		0,1,2,3,4,5,6,7,8,9:
			instance.set_event_type("process")
			instance.add_to_group("processes")
		10:
			instance.set_event_type("connection")
			instance.add_to_group("networks")

	instance.connect("exploded", _on_projectile_exploded)
	add_child(instance)

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
