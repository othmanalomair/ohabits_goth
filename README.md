  # On your VPS, navigate to the app directory
  cd ~/apps/ohabits_goth

  # Pull the latest changes from GitHub
  git pull origin main

  # Rebuild the Docker image with the updated code
  docker build -t apps_app:v2.0.1 .

  # Update the Docker service to use the new image
  docker service update --image apps_app:v2.0.1 apps_app

  # Check if the service is now running properly
  docker service ps apps_app

  # Check the logs to make sure it's working
  docker service logs apps_app
