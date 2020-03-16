#!/bin/bash 

set -e


if [ $# -lt 2 ] ; then
    echo "Usage: ./get_version.sh <JOB> <BUILD_NUMBER>";
    exit 1
fi

JOB=$1
BUILD_NUMBER=$2


# Function to parse yaml file
# The yaml values are printed as k1=v1\nk2=v2\n...
# The values can be pulled into bash variables (with an optional prefix) by running with eval
# For example, running eval $(parse_yaml some.yaml PREFIX_) will pull the yaml values into bash variables, each with the prefix PREFIX_
function parse_yaml {
   local prefix=$2
   local s='[[:space:]]*' w='[a-zA-Z0-9_]*' fs=$(echo @|tr @ '\034')
   sed -ne "s|^\($s\):|\1|" \
        -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p"  $1 |
   awk -F$fs '{
      indent = length($1)/2;
      vname[indent] = $2;
      for (i in vname) {if (i > indent) {delete vname[i]}}
      if (length($3) > 0) {
         vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
         printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
      }
   }'
}

# Function to get GIT workspace root location
function get_git_ws {
    git_ws=$(git rev-parse --show-toplevel)
    [ -z "$git_ws" ] && echo "Couldn't find git workspace root" && exit 1
    echo $git_ws
}

# Pull the major, minor, maintenance versions from the repository's version.yaml file
version_file=$(get_git_ws)/version.yaml
eval $(parse_yaml $version_file AKO_VERSION_)

# Compute base_build_num
base_build_num=$(cat $(get_git_ws)/base_build_num)
version_build_num=$(expr "$base_build_num" + "$BUILD_NUMBER")

version_tag="$AKO_VERSION_major.$AKO_VERSION_minor.$AKO_VERSION_maintenance-$version_build_num"

mkdir -p /tmp/$JOB;
touch /tmp/$JOB/jenkins.properties;
echo "version_tag=${version_tag}" > /tmp/$JOB/jenkins.properties;
echo $version_tag
