/*
Package coordinator implements all features necessary to initialize and monitor elasticsearch components

On-start :

* LoadConnectors Configuration :
  - Instances + Logical Indices + Cron

* Check if any corresponding indices exists
* If so, Check if any corresponding aliases exists
* Check if any corresponding templates exists
*/
package coordinator
