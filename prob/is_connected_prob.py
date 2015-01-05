#!/bin/python2

from scipy.stats import binom

# number of connections
c1 = 8 
c2 = 75
c3 = 75

# probabilities for the (conditional) binomial variables
p1 = float(4)/(c1*c2)
p2 = float(8)/(c1*c2*c3)


# optimize for n
for n in range(5000):
    k = 0
    for k in range(n):
        B1 = binom.cdf(k, n, p1)
        B2 = binom.cdf(k, n, p2)
        if B1 > 0.01:
            break
        if B1 < 0.01 and B2 > 0.99:
            print "solution:", k, n
            exit(0)
print "found no solution"
