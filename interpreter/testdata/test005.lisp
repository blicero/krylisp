;; /home/krylon/go/src/krylisp/interpreter/testdata/test005.lisp
;; created on 06. 11. 2017
;; (c) 2017 Benjamin Walkenhorst
;; Time-stamp: <2017-11-10 20:49:50 krylon>
;;
;; Montag, 06. 11. 2017, 21:32
;; So, now we're on to loops. I *could* just copy Common Lisp's DO special form,
;; which I like. Or I could think up something of my own.
;; With Common Lisp, the situation is kind of strange, because the first time 'round,
;; LOOP was very primitive, basically just a placeholder. On the second iteration, however,
;; LOOP had evolved into a very... complex thing.
;; I do not want that.
;; I think replicating the DO loop will be sufficient, as that one is fairly powerful
;; to begin with.
;; But then I need a generalized interface between the DO-loop and the bunch of stuff that
;; going to get processed. Kind of like Python's Iterator protocol. *NOT* like C++'s
;; iterators (I like them, asthetically, but I would not like to have to implement them).


;; Freitag, 10. 11. 2017, 20:49
;; XXX For now, this script is unused. I might come back later to write test
;;     using this script, but for now it is just dead code.

(define measures [ [ 5.5 7.3 8.1]['Peter 'Karl 'Horst]])

(define performance-log { 1987: [ 12 21 19 37 45 ],
                          1988: [ 17 4 9 8 64],
                          1989: [ 32 64 96 128 256 ] })





