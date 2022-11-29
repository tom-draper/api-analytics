from django.contrib import admin
from django.urls import path, include
from .views import say_hello
 
urlpatterns = [
    path('', say_hello, name='say_hello'),
]