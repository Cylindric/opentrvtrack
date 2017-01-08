from rest_framework import viewsets
from django.contrib.auth.models import User
from .models import Sensor
from .serializers import SensorSerializer

class SensorViewSet(viewsets.ModelViewSet):
	serializer_class = SensorSerializer
	queryset = Sensor.objects.all()
