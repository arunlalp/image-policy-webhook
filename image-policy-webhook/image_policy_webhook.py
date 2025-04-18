from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/validate', methods=['POST'])
def validate_image():
    allowed_registries = ['registry.example.com', 'docker.io']

    admission_review = request.get_json()
    pod = admission_review['request']['object']

    for container in pod['spec']['containers']:
        image = container['image']
        registry = image.split('/')[0]

        if registry not in allowed_registries:
            return jsonify({
                'apiVersion': 'admission.k8s.io/v1',
                'kind': 'AdmissionReview',
                'response': {
                    'uid': admission_review['request']['uid'],
                    'allowed': False,
                    'status': {
                        'message': f"Image {image} is not allowed. Allowed registries: {', '.join(allowed_registries)}"
                    }
                }
            })

    return jsonify({
        'apiVersion': 'admission.k8s.io/v1',
        'kind': 'AdmissionReview',
        'response': {
            'uid': admission_review['request']['uid'],
            'allowed': True
        }
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080, ssl_context=('/app/certs/tls.crt', '/app/certs/tls.key'))