aliker
======

*Find similar things on tumblr*

## Conjecture
The things that you're likely to like are the things that are liked by the people who like the things that you like.

## Algorithm
* Initialize an empty map of posts to a list of users: `M`
* Accept post from some blog: `P`
* Find every user who has left a note on that post (either a like or reblog) `U0` - `Un`
* For each user `U0`-`Un`: `U`
  * For each note that `U` wrote to some post `Q`:
    * Append `U` to the list `M[Q]`
* (TODO: somehow normalize this map so that some peoples' counts mean more than others -- not yet implemented)
* Initialize a second empty map of posts `Q` to counts of users liking each post: `C`
* For each key `Q` in `M`:
  * Set `C[Q]` to the length of the list in `M[Q]`
* Now, the largest values in `C` have the most similar keys (posts) to `P`

## Implementation
It's in Go. It uses websockets, bootstrap, jQuery and the gorilla web toolkit.
