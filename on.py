#library imports
import RPi.GPIO as GPIO
import time
# defines the numbering scheme the pins use
GPIO.setmode(GPIO.BCM)
GPIO.setup(17, GPIO.OUT)
GPIO.output(17,GPIO.HIGH)
GPIO.cleanup()

