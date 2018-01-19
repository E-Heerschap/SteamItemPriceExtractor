#Introduction
The steam item manager is a package of tools which handles the gathering, processing and storing of information for steam items. The package currently consists of two tools:

* HashNameExtractor

>To get information on a steam item steam api requires the *hash name* of the item. The HashNameExtractor downloads and stores a list of all the hash names of items currently on the steam market.

* PriceExtractor

>PriceExtractor is self explanatory. It uses the the hash names extracted using the *HashNameExtractor* to download information on the item.

#Dependencies
This package requires the following dependencies to compile and run correctly.

###x/net/html
This is the package used to scrape the html from the steam responses.

####Installation:
    $ go get golang.org/x/net/html/

The documentation can be found on [go docs](https://godoc.org/golang.org/x/net/html).

###go-sql-driver
This is the package that lets golang interact with the mysql server.

####Installation:

    $ go get github.com/go-sql-driver/mysql

The documentation can be found on its [github page](https://github.com/go-sql-driver/mysql).

###Tor
Tor allows the the *HashNameExtractor* and *PriceExtractor* switch the IP address that is appearing to steam to accelerate our request speeds. Tor needs to be running on the server(s) that will be running this package.

####Installation:
    $ sudo apt-get update
    $ sudo apt-get -y install Tor

To test if the Tor service started check it by typing:

    $ service Tor status

Next the Tor config needs to be changed to allow interaction with the control port.

Create a password, for the Tor control port, that you will put into the Tor config. Tor hashes your password using the following command:

    $ tor --hash-password PASSWORD

Copy the password to your clipboard. The password will look similar to this:

16:929101F062ED0A6D60D16977B3215272C65455FD9B3A1E76D826E57840

Next edit the configuration file (torrc) found at */etc/tor/torrc*. Edit this file using your preferred text editor such as nano or vim and sudo.

    $ cd /etc/tor/
    $ sudo nano torrc

Remove the comment, # symbol, infront of the following line:

*#ControlPort 9051*

and remote the comment and replace the existing hash on the line:

*#HashedControlPassword 16:872860B76453A77D60CA2BB8C1A7042072093276A3D701AD68405D684053EC4C*

Then save the file and restart tor.

    $ sudo service tor restart


#TODO ON README

* Explain configuration files

* Explain which directories to compile each of the seperate packages in.

* Explain database setup


##Notes:

Database VARCHAR fields must use UTF-8 encoding to work correctly (Otherwise we cannot store csgo items with fancy names).

## Useful Links and Information

[Get steam market hash names](https://www.reddit.com/r/SteamBot/comments/2v05by/identifying_every_item_on_the_market/)

[Get steam app ip list](http://api.steampowered.com/ISteamApps/GetAppList/v0001)

[Get MySql Driver](https://github.com/go-sql-driver/mysql)

[Get top few item listings](https://stackoverflow.com/questions/26513891/get-steam-item-prices)
