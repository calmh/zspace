zspace
======

Shows space used nicely categorized by dataset type and "class", where
class is user defined by regexp. Requires a recent illumos flavoured zfs
binary.

```
[root@anto ~]# zspace 
CATEGORY             DATASET    SNAPSHOT      REFRES       TOTAL  COMPRESS
class/backup         94.3 GB     80.3 GB      0.0 GB    174.6 GB     1.87x
class/dl           1709.7 GB      1.6 GB      0.0 GB   1711.3 GB     1.00x
class/foto          319.0 GB      1.0 GB      0.0 GB    320.0 GB     1.02x
class/git             0.5 GB      0.0 GB      0.0 GB      0.5 GB     0.52x
class/music         132.6 GB      0.0 GB      0.0 GB    132.6 GB     1.01x
class/other           0.0 GB      0.0 GB      0.0 GB      0.0 GB     1.29x
class/system         11.5 GB      2.5 GB     15.6 GB     29.6 GB     1.89x
class/vm            161.3 GB    255.3 GB    297.0 GB    713.7 GB     1.22x
type/filesystem    2282.8 GB     86.6 GB      0.0 GB   2369.4 GB     1.05x
type/volume         146.1 GB    254.2 GB    312.6 GB    712.8 GB     1.25x
TOTAL              2428.9 GB    340.7 GB    312.6 GB   3082.2 GB     1.06x
```

Install
-------

Drop the binary in `/opt/local/bin`, drop `zspace-classes.txt` in
`/opt/local/etc`. Edit the latter to suite; the format is
`<name><space><regexp>`.

Run
---

Run `zspace`. See `zspace -help` for options.


License
-------

MIT

