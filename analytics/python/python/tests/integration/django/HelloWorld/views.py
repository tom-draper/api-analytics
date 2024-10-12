from django.http import JsonResponse

# Create your views here.
def say_hello(request):
    return JsonResponse({'message': 'Hello World!'})
