#!/usr/bin/env python
import sys, yaml

with open(sys.argv[1]) as f:
    data = yaml.load(f)
    for dep in data['dependencies']:
        repo = dep['repository'].replace("file://", "")
        print(repo)
