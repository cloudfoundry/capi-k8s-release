[ ] remove `config` directory and any related make directives
[ ] have CI manage ensuring that version of kpack CRDs matches the version of
kpack source we use
    - watches github releases of kpack: new release => bump kpack version in
      go.mod and copy in the new release yml into the `tmp` dir of this repo
