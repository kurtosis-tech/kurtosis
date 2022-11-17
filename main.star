service = add_service(service_id = "httpd-service", config = struct(image = "httpd:2.4.54", ports = {"http" : struct(number = 80, protocol = "TCP" )}))
print("httpd has been added successfully")