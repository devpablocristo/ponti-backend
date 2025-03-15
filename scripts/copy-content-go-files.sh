#!/bin/bash

find . -name "*.go" -type f -exec cat {} + | xclip -selection clipboard¿