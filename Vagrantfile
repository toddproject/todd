# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
 
  config.vm.box = "trusty64"
  config.vm.box_url = "http://cloud-images.ubuntu.com/vagrant/trusty/current/trusty-server-cloudimg-amd64-vagrant-disk1.box"
  config.vm.synced_folder '.', '/home/vagrant/go/src/github.com/Mierdin/todd', nfs: true

  # This is a temporary measure. Once https://github.com/Mierdin/todd/issues/72 is addressed, we can go back
  # to a single folder being redirected (namely toddproject). For now, since ToDD is still in "Mierdin",
  # we need to do it this way.
  config.vm.synced_folder '../../toddproject/todd-nativetestlet-ping', '/home/vagrant/go/src/github.com/toddproject/todd-nativetestlet-ping', nfs: true

  config.vm.provider "virtualbox" do |v|
    v.memory = 2048
    v.cpus = 2
  end

  config.vm.define "todddev" do |todddev|

    todddev.vm.host_name = "todddev"
    todddev.vm.network "private_network", type: "dhcp"
    todddev.vm.network "public_network", type: "dhcp"
    todddev.vm.provision "ansible" do |ansible|
        ansible.playbook = "scripts/ansible-todd-dev.yml"
    end
  end
        
end
