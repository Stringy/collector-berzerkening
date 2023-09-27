extends Control

@export var process_score: int = 1
@export var conn_score: int = 5
@export var endpoint_score: int = 10

# Called when the node enters the scene tree for the first time.
func _ready():
	pass # Replace with function body.


# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	pass

func process():
	$Board/Processes.add()
	update_score(process_score)
	
func connection():
	$Board/Connections.add()
	update_score(conn_score)
	
func endpoint():
	$Board/Endpoints.add()
	update_score(endpoint_score)
	
func update_score(value):
	$Board/Score.add(value)
