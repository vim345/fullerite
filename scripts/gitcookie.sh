touch ~/.gitcookies
chmod 0600 ~/.gitcookies

git config --global http.cookiefile ~/.gitcookies

tr , \\t <<\__END__ >>~/.gitcookies
.googlesource.com,TRUE,/,TRUE,2147483647,o,git-fdc.yelp.com=1/xWVRFqV6Fcudb7yMZezytnNHYsKDfQ9KFwAcLiV5w3g
__END__
