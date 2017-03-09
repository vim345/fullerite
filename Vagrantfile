# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
# -*- ruby -*-
VAGRANTFILE_API_VERSION = "2"

stub_name = "fullerite"

boxes = {
  :trusty  => 'ubuntu/trusty64',
  :precise => 'puppetlabs/ubuntu-12.04-64-nocm',
  :lucid   => 'chef/ubuntu-10.04',
}

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  # 192.168.33.210-229
  nodes = [
    { name: 'fullerite-vm1', memory: '512', box: 'trusty',  master: true, },
  ]

  go_tgz = "go1.7.3.linux-amd64.tar.gz"

  if Vagrant.has_plugin?("vagrant-cachier")
    config.cache.scope = :box
  end

  nodes.each do |node|
    config.vm.define "#{node[:name]}.#{stub_name}" do |vm_config|

      vm_config.vm.box = boxes[node[:box].to_sym]
      vm_config.vm.hostname = "#{node[:name]}.#{stub_name}"

      if ! node.fetch(:ip, nil).nil?
        vm_config.vm.network :private_network, ip: node[:ip]
      end

      vm_config.vm.provider "virtualbox" do |v|
        v.customize ["modifyvm", :id, "--memory", node[:memory]]
        v.name = node[:name]
      end

      node.fetch(:sync_dirs, []).each do |sync_dir|
        puts sync_dir.fetch(:source)
        puts sync_dir.fetch(:dest)
        vm_config.vm.synced_folder sync_dir.fetch(:source), sync_dir.fetch(:dest)
      end

      vm_config.vm.provision "shell", inline: "apt-get update && apt-get -y dist-upgrade && apt-get install -y git python-pip"
      vm_config.vm.provision "shell", privileged: true, inline: "[ ! -f #{go_tgz} ] && wget -q https://storage.googleapis.com/golang/#{go_tgz} && tar -C /usr/local -xzf #{go_tgz}"
      vm_config.vm.provision "shell", inline: "echo 'PATH=/usr/local/go/bin:$PATH' >> .profile"
      vm_config.vm.provision "shell", inline: "pip install --user -r /vagrant/requirements-dev.txt"
      vm_config.vm.provision "shell", privileged: true, inline: "mkdir /etc/fullerite && cp /vagrant/vagrant/fullerite.conf /etc/fullerite.conf && rsync -r /vagrant/vagrant/conf.d/ /etc/fullerite/conf.d/"

    end

  end

end
