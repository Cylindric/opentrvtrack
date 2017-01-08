from django.shortcuts import get_object_or_404, render

from .models import Sensor

def index(request):
	sensor_list = Sensor.objects.order_by('name')[:5]
	context = {
		'sensor_list': sensor_list,
	}
	return render(request, 'sensors/index.html', context)


def sensor(request, sensor_id):
	sensor = get_object_or_404(Sensor, pk=sensor_id)
	return render(request, 'sensors/detail.html', {'sensor': sensor})

