from django.conf.urls import url

from . import views

app_name = 'dashboard'

urlpatterns = [
	# /sensors/
	url(r'^$', views.index, name='index'),

	# /sensors/1
	url(r'^(?P<sensor_id>[0-9]+)/$', views.sensor, name="sensor"),
]

