
'use strict';

var path = require('path');
var fs = require('fs');
var util = require('util');

var utils = require('fabric-client/lib/utils.js');
var Peer = require('fabric-client/lib/Peer.js');

var config = require('../config.json')
var helper = require('./helper.js');
var logger = helper.getLogger('install-chaincode');

var tx_id = null;
var nonce = null;
var adminUser = null;

//function installChaincode(org) {
var installChaincode = function (peers, chaincodeName, chaincodePath, chaincodeVersion, username, org){
    logger.debug('\n============ Install chaincode on organizations ============\n')
		helper.setupChaincodeDeploy();
		var chain = helper.getChainForOrg(org);
		helper.setupOrderer();
		var targets = helper.getTargets(peers, org);
    helper.setupPeers(chain, peers, targets);
		/*for(var index in targets) {
			chain.addPeer(targets[index]);
		}

    logger.info(chain.getPeers());*/

		return helper.getRegisteredUsers(username,org)
		.then((member) => {
		adminUser = member;

		nonce = utils.getNonce();
		tx_id = chain.buildTransactionID(nonce, adminUser);

		// send proposal to endorser
		var request = {
			targets: targets,
			chaincodePath: chaincodePath,
			chaincodeId: chaincodeName,
			chaincodeVersion: chaincodeVersion,
			txId: tx_id,
			nonce: nonce
		};

		return chain.sendInstallProposal(request);
	},
	(err) => {
		logger.error('Failed to enroll user \'admin\'. ' + err);
		throw new Error('Failed to enroll user \'admin\'. ' + err);
	}).then((results) => {
		var proposalResponses = results[0];

		var proposal = results[1];
		var header   = results[2];
		var all_good = true;
		for(var i in proposalResponses) {
			let one_good = false;
			if (proposalResponses && proposalResponses[0].response && proposalResponses[0].response.status === 200) {
				one_good = true;
				logger.info('install proposal was good');
			} else {
				logger.error('install proposal was bad');
			}
			all_good = all_good & one_good;
		}
		if (all_good) {
			logger.info(util.format('Successfully sent install Proposal and received ProposalResponse: Status - %s', proposalResponses[0].response.status));
			logger.debug('\nSuccessfully Installed chaincode on organization '+org+'\n')
			return 'Successfully Installed chaincode on organization '+org;
		} else {
			logger.error('Failed to send install Proposal or receive valid response. Response null or status is not 200. exiting...');
			return 'Failed to send install Proposal or receive valid response. Response null or status is not 200. exiting...';
		}
	},
	(err) => {
		logger.error('Failed to send install proposal due to error: ' + err.stack ? err.stack : err);
		throw new Error('Failed to send install proposal due to error: ' + err.stack ? err.stack : err);
	});
}
exports.installChaincode = installChaincode;
