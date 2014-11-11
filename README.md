# Mac OS X kernel tracer

When the Mac OS X Mach kernel starts vomiting uncontrollably, dtrace is the tool of choice for getting a sniff of what's going on. However, when things actually do go wrong, the last thing you want to need doing is to be typing in dtrace commands from your phone. Based on Brendan Gregg's excellent [hotkernel](http://www.brendangregg.com/DTrace/hotkernel), `osxkerneltracer` makes it easy to get a sampling call trace dump from your kernel.

## Building and running

`osxkerneltracer` is written in Go and therefore needs the Go toolchain to compile. If you're any kind of power user, [Homebrew](http://brew.sh/) will get you there in a jiffy. Building and installing the application is then a simple call to `make` away:

    $ make all install

Getting a trace is now simple when your system goes AWOL, but remember that root privileges are required for tracing the kernel:

    $ sudo osxkerneltracer

The output will look something like the following:

    Module | Method                     | Calls |      Share
    -------+----------------------------+-------+-----------
    kernel | machine_idle+0x1fd         | 72179 |  98.5271 %
    kernel | 0xffffff8013552880+0x313   |   349 |   0.4764 %
    vmmon  | Task_Switch+0xed1          |   169 |   0.2307 %
    kernel | processor_idle+0xc5        |   127 |   0.1734 %
    kernel | processor_idle+0xb4        |    30 |   0.0410 %
    kernel | ipc_mqueue_post+0x227      |    22 |   0.0300 %
    kernel | 0xffffff8013552880+0x3bf   |    18 |   0.0246 %
    kernel | machine_idle+0x1ff         |    11 |   0.0150 %
    kernel | thread_block_reason+0xbe   |    10 |   0.0137 %
    kernel | lck_mtx_lock_spin+0x28     |     9 |   0.0123 %
    kernel | processor_idle+0xab        |     8 |   0.0109 %
    kernel | thread_continue+0x43       |     8 |   0.0109 %
    kernel | vnode_iterate+0x1e9        |     8 |   0.0109 %
    kernel | processor_idle+0xda        |     7 |   0.0096 %
    vmmon  | Task_Switch+0xf74          |     7 |   0.0096 %
    kernel | hfs_lock+0x1f              |     6 |   0.0082 %
    kernel | processor_idle+0xdf        |     6 |   0.0082 %
    kernel | thread_call_enter1+0x2b5   |     6 |   0.0082 %
    kernel | lck_rw_lock_exclusive+0x24 |     5 |   0.0068 %
    kernel | processor_idle+0xbb        |     5 |   0.0068 %
    kernel | wait_queue_wakeup_all+0xb9 |     5 |   0.0068 %
    ..

If you want to trace a duration other than the default 5 seconds, pass the duration as a flag argument:

    $ sudo osxkerneltracer -d 10s
