# wemo-insightmonitor
wemo insights power monitor

This monitor will get the power measurement every 10 seconds.
if the power is above a certain threshold, we assume the washingmachine is on.
If the power is below a certain threshold, we assume the washingmachine is finsihed.

All message are broadcasted to a nsq message bus for further handling by your application.

The topic is: wemo_monitor_washingmachine

The ip address of your device must be changed to your needs.


