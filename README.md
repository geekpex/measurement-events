# measurement-events

This tool reads measurement info from STDIN and writes warning events to STDOUT.

Supported input format:
```
Time,Level
1,0
2,1
4,1
5,0
```


Output format:
```
Start Time,End Time,Level
2,4,1
```


## Usage
```
	-k	--keep-running	Exit only when interrupt or terminate is received
```