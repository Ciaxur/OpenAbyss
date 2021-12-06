#!/usr/bin/env bash

SCRIPT_DIR=`dirname $0`
pushd "$SCRIPT_DIR/.." 1> /dev/null
mkdir build 2> /dev/null

platforms=( 'darwin' 'linux' 'windows')
archs=('amd64' 'arm64')
# GOOS=darwin GOARCH=amd64 go build -o hello_world_macOS

# Release Build
if [ "$1" == "--release" ]; then
  for platform in "${platforms[@]}"; do
    for arch in "${archs[@]}"; do
      # Build client binary for declared archs & platforms
      client_binary_name="client-$platform-$arch"
      echo "Building to '$client_binary_name'"
      GOOS=$platform GOARCH=$arch go build -o build/$client_binary_name ./client

      # Build client binary for declared archs & platforms
      server_binary_name="server-$platform-$arch"
      echo "Building to '$server_binary_name'"
      GOOS=$platform GOARCH=$arch go build -o build/$server_binary_name ./server
    done
  done

elif [ "$1" == "--install" ]; then # Installs client for Linux Users
  echo "Installing open-abyss client binary..."
  INSTALL_PATH="/opt/OpenAbyss"
  CLIENT_CONFIG=".config/config-client.json"

  # Create path for OpenAbyss if fresh install, otherwise
  #  remove previous binary.
  echo "Creating OpenAbyss User Directory '$INSTALL_PATH'"
  if [ -d $INSTALL_PATH ]; then
    rm $INSTALL_PATH/open-abyss 2> /dev/null
  else
    mkdir $INSTALL_PATH 2> /dev/null
  fi

  # Check for faults
  if [ $? != 0 ]; then
    echo "Run installation process with sudo." && exit 1
  fi

  # Build Client binary
  echo "Building client binary."
  go build -o $INSTALL_PATH/open-abyss ./client

  # Check if certificates were created and warn user of certificates not
  #  being present, which would require another install if certificates
  #  are generated after installing the binary.
  if [ -d ./cert ]; then
    echo "Installing certificates used to communicate with the server."
    cp -r ./cert $INSTALL_PATH/
  else
    echo "WARN: No certificates found. If you generate SSL certificates " \
      "after installing this binary, be sure to re-run the installation " \
      "process to ensure certificates are copied over."
  fi

  # Copy over configurations if present in current directory and
  #  not in install path.
  if [ -f $CLIENT_CONFIG ] && ! [ -d $INSTALL_PATH/.config ]; then
    echo "Installing client configurations."
    mkdir $INSTALL_PATH/.config 2> /dev/null
    cp $CLIENT_CONFIG $INSTALL_PATH/.config/
  fi

  # Create the final piece: The symlink!
  if ! [ -f /usr/bin/open-abyss ]; then
    echo "Installing symlink within the user's '/usr/bin' path"
    ln -s $INSTALL_PATH/open-abyss /usr/bin/
  fi

  # Final CLI-bliss
  echo -e "\nTo enable shell completion, add the following to your shell source rc,"
  echo 'For bash: eval "$(open-abyss --completion-script-bash)"'
  echo 'For zsh: eval "$(open-abyss --completion-script-zsh)"'

else
  # Assumes "local" testing, with respect to the cloned repo's path
  # Build based on current platform/architecture
  go build -o build/server ./server
  go build -o build/client ./client

  # Copy certificates, if available, into build path for client to use
  if [ -d ./cert ]; then
    echo "Copying certificate directory to '$SCRIPT_DIR'"
    cp -r ./cert ./build/
  fi

fi

popd 1> /dev/null