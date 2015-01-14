#!/bin/python2

from scipy.stats import binom
from scipy.special import binom as binom_coef
import math

# Compute the probability that two nodes are connected after sending n addr messages
# and receiving k messages back

# Events: C the two nodes are connected connected; received k addr messages
# Use Bayes' theorem
# P(C| k) = (P(k | C) * P(C))/(P(k and C) + P(k and not C)
# P(k | C) = binom.cdf(k, n, p1) 
# P(C) = 1/(num_nodes over 2)
# P(k and C) = P(k | C) * P(C)
# P(k and not C) = P(k | not(C)) * P(not(C)) > binom.cdf(k, n, p2) * 1-P(C)
# P(C | k) = 1 - P(nC | k)

# number of connections
c_fullnode = 60
c_client = 8

c1 = c_client
c2 = c_fullnode
c3 = c_fullnode
# number of full nodes in total in the network
numNodes = 6793
# total number of addr messages sent
n = 2000
# number of addr messages received
k = 8

# probability that node 1 relays to node 2 and node 2 relays to attacker
p1 = float(4)/(c1*c2)
# probability that node 1 relays to node 2 and node 2 relays to node 3
# and node 3 relays to attacker
p2 = float(8)/(c1*c2*c3)


# probability that k addr messages are received given the nodes are connected
PkC = binom.pmf(k, n, p1)
# upper bound probability that two nodes are connected if they are only allowed to connect to one peer
# basically we are computing the probability that c1 is not not selected as an outbound peer of c2 
# and vice versa eventually
PnC = reduce(lambda x, y: (1-(1.0/y))*x, range(numNodes-8, numNodes), 1)
PC = 1-PnC
PkandC = PkC * PC
PnC = 1- PC
# upper bound on the probability that k addr messages are received given the nodes are not connected
PknC = binom.pmf(k, n, p2)
PkandnC = PknC * PnC

# probability that two nodes are connected given k addr messages were received
p = PkandC/(PkandC + PkandnC)
print "PkC", PkC, "PC", PC, "PkandC", PkandC, "PnC", PnC, "PknC", PknC, "PkandnC", PkandnC 
print "probability that these nodes are connected:", p
print "upper bound on probability that more than k messages are returned given that they are not connected:",  1 - binom.cdf(k, n, p2) 
print "probability that k messages or less are returned given that they are connected:",  binom.cdf(k, n, p1) 
