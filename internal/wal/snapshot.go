/* TODO
* modify the wal record to have ids for each operation
- implement getCurrState function in memory.go
- find a way to fire state logging function periodically or on POST clear operation
- write init state logger function 
- setup state logger
    - getCurrState()
    - get current operation index (walCheckpoint)
    - construct stateRecord
    - append stateRecord to state.json
- modify wal replay
	- fetch latest state
	- extract walCheckpoint
	- open wal.log and iterate till you get the walCheckpoint
	- and execute operations *after* that checkpoint
*/
package wal