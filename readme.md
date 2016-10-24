entities
==
task
  - name
  - script

run
  - id
  - task-id
  - vc-ref
  - status
  - stdout
  - stderr

interface
==

```
POST /tasks {"name": "test cd", "script": "export GOPATH=/home/feller/\ngit clone https://github.com/fgeller/contd\ncd contd\ngo test -v\n"}
 -> 200 {"id": "abc"}
```

```
POST /tasks/abc/runs {"task-id": "abc", "vc-ref": "cafecafe"}
 -> 200 {"id": "24"}
```

```
GET /tasks
 -> 200 {"tasks": [{"id": "abc", "name": "test cd", "script": "export GOPATH=/home/feller/\ngit clone https://github.com/fgeller/cd\ncd cd\ngo test -v\n", "runs": [{"id": "24", "task-id": "abc", "vc-ref": "cafecafe", "status": "pending", "stdout": null, "stderr": null}]}]}
```

BONUS

```
GET /tasks/abc/runs/24/out Connection: upgrade
 -> 200 websocket
```
