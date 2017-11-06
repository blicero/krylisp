;; /home/krylon/go/src/krylisp/interpreter/testdata/test004.lisp
;; created on 03. 11. 2017
;; (c) 2017 Benjamin Walkenhorst
;; Time-stamp: <2017-11-06 19:46:50 krylon>

;; I just added regular expressions to the type set, and now I need to figure
;; out what I want the API to look like.
;;
;; I have avoided handling escape sequences in strings, yet.
;; I am not sure if I want to.
;; But with regexps, this becomes relevant.
;; I would like to have syntax for regexp literals.

(define pattern (regexp-compile "(?im)^(\d+)\s+(\w+)$"))

(let ((res (regexp-match pattern "1234     Peter")))
  (print (aref res 0))
  res)



