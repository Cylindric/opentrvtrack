from __future__ import unicode_literals

from django.db import models

# Create your models here.

class Sensor(models.Model):
	serial = models.CharField(max_length=16)
	name = models.CharField(max_length=100)
	
	def __str__(self):
		return self.serial

