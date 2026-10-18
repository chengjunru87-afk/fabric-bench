'use strict';
const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class CredWorkload extends WorkloadModuleBase {
    constructor() { super(); this.i = 0; }
    async initializeWorkloadModule(w, tw, r, args, sut, ctx) {
        await super.initializeWorkloadModule(w, tw, r, args, sut, ctx);
        this.cid = args.contractId;
        this.op = args.operation;
    }
    async submitTransaction() {
        this.i++;
        if (this.op === 'issueCredential') {
            const id = 'bench-' + this.workerIndex + '-' + this.i;
            return this.sutAdapter.sendRequests({
                contractId: this.cid,
                contractFunction: 'IssueCredential',
                invokerIdentity: 'Admin',
                readOnly: false,
                contractArguments: [id, 'did:fabric:china:chesicc:issuer123', 'did:kyrgyz:keu:stu' + this.i, 'bachelor', 'Engineering', 'XJU', '2024-06-30', 'doc' + this.i]
            });
        }
        return this.sutAdapter.sendRequests({
            contractId: this.cid,
            contractFunction: 'VerifyCredential',
            invokerIdentity: 'Admin',
            readOnly: true,
            contractArguments: ['cred-001']
        });
    }
}
module.exports.createWorkloadModule = () => new CredWorkload();
