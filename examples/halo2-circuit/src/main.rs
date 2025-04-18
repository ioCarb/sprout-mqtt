pub mod opts;

use std::{fs, io::BufReader};

use clap::Parser;
use halo2_simple_circuit::{
    circuits::simple::SimpleCircuit,
    generator::{gen_pk, gen_proof, gen_sol_verifier},
};
use halo2_curves::bn256::{Bn256, Fr, G1Affine};
use halo2_proofs::{
    circuit::Value,
    plonk::verify_proof,
    poly::{
        commitment::{Params, ParamsProver},
        kzg::{commitment::ParamsKZG, multiopen::VerifierGWC, strategy::AccumulatorStrategy},
        VerificationStrategy,
    },
    transcript::TranscriptReadBuffer,
};
use itertools::Itertools;
use opts::{Opts, Subcommands};
use serde_derive::{Deserialize, Serialize};
use snark_verifier::{loader::evm::{self, encode_calldata}, system::halo2::transcript::evm::EvmTranscript};

#[derive(Debug, Deserialize, Serialize)]
struct Proof {
    proof: String,
    calldata: String,
}

fn main() {
    let opts = Opts::parse();

    // TODO replace your params
    // invoke gen_srs() to get your params
    let params_hex = "040000009d0d8fc58d435dd33d0bc7f528eb780a2c4679786fa36e662fdf079ac1770a0e3a1b1e8b1b87baa67b168eeb51d6f114588cf2f0de46ddcc5ebe0f3483ef141c14188466fc341ccc92e7dd5b129957750613f1989455251f8c847ab6ebb9d917b65f30ca035be6fdd7fb8cd72f9ea00845b63997ef8e3f58eeff97a93803cc0305ce30bfc51599e87657226910d9b8def42a09ebb9460d5048aca2601cd39419c7290510857909131a7137030819a47d43dfb81b4837be80554367e4bfea8f03796b7f2a7e10d20e216d0b5a579b2287d12c575071571cb4be698d15686b28287cfb767a933826899b9c929dcc1b9cc34722b7c65d2a9569f5914229643dd30beaadf1cae90d24bc0d6157ef3555c843f7e4d12c4ae71adb56ecd09b5156fa1353c06dfa727cf8a18e3d7ae52b8a6ccc6df696cb2732a05d0a3a3fa355cc05017dafb68ce091f9821f6996cd5df79d874c31fdea4e1bf75a75f9e1d57d32d009b76e2e848004c4a7be55d1bef5bef6cd584b5b00b758e199999a3ea4c811b722de4c1c404293388017c2a5e1369640ef9b6437b44fd066406a72419e0233e325ac5916b95b69f7f53f29f4a0bb163f442bebf009f06e0313b04296a4e2fa330da005e9d20031f5f48ec49718cb2a7ed2c19e7e525bd9c7867314b9c72226f02530137fb4fd113e3cc907a2d863a3df5e6abd4eb7433c79386b69df7867fdf80052599411b7b4c4957a8c9d34beca2820c89ff7c5ee57d060f8517d200bc0fc22b671964cd8a0b84d39af206908c3ccd323984991e95881f08ecd35f0583dc90acff5164cc7724f25ebd71d9856c06df71c2867d8df5b6f20537497b7899d7d2d61e808bd7e56f301de7875919ec379c2f43e02b3a3ba5c17edbe7e6db6f49300f5831b1068c0f83b1f7ed9c2a3173dd60bd8aa58bee555681c72aabce0350502aa1b125b6e037ca209c82a2c22ee946e6d80aff3b3d57055b7e4ab0c472b9d0d4d415629137bb08a89748b1b5be4359ad562956c55f8aa061f5f33987f9ef013c04e6cf3caf1aea02c8836c3d92caa56cebece09cdfd9807a015081777dd0a2f2ce81d0c559a72d212e7e8afdac7dfedf5a2daf76561038fd9691dd2f17e142368d171edaaf21dba842951d4af8225c64a0f701b0e2f64064dc1f4aed285fa26ec5fe1a233f54851e0d81a9afa1d8626326ec5bc5ce8bb6c34972e53e8c7dc0a7fd2d019a852c64017872cd0d40a3b9abb9327b403ece3c60a5831a3e68f9d0293043d0ebe801d752fb1a5709d0e71f9b95a93694af0719c845328d598773f07ca425f9fc1447af7173678270e3b95ff34ce2a93e8e401ad40369c4c728b430bd0c2979deb2a89ff8284588790a93cbda5e85f5077d5443b00a1a849afb28d149c9468240475fff6edcc58a8aff85f9b99f2ccc687fd5fec024d95981cca2615d984bfe7833ee40e4a0ad3035360b59cf2fe8f4b00ddcb1a7e7f4d6e7f3c7b13a668eade3581d08270b61b184af720e8fdf5370db6000161ba0d9a8c4e8953048a1816f09ba747572703a4242e74a038a0bda0d29a7a645cdb7ca6c534b51423716294d3a87cc1c5831c135c0dc005d8e09e9685fc54a8ebbcc448331bf41b2a090e7d571ad654adb77a88488bf3a330e582f30e416e785ee8d04050299ca027e73f8737d3a43a67d7b775f792bd96bbe036e71283a75894c9c5cf65bf9eba2f15f7120f1fdc371b8739ddf97a07f6a9dc199a326acab6b9eed7369337ba4d03084177bf90bd48b2233eae2ecf6c32ac18fed417ed492daf4afd0ae8fd90af16dbbe55af4a53e5f52f7c5f02b817f9386dbcbeadeed835fb2048c6c11ee1a51f6696d0d95a14894de67a51647500a9d928119f941a82f8e2e53f9986ebebd710460dd51167cef34b696b74483f32d807604decb8a247e4cfe6c6c9d3af40582999055e68fef94fc12ea5c186c0d057bc0726f79b270065cf44ae94ac55837616cbc1abb7a199f1e8b4e314d0bfd071c5e2ab250c91dd7b32c70efcb4eff1bf0ed3c8c635a0ce5a75a4efab8894477f33647ebea2d370953923af6174514e0b1c3234e05c7662cdf1ac0846701eeabe83316a0a6cbae79ab4bbc50e6a0603f72eb34ab58db62e6e6b09b4293a9da90096a90717a46cb77481af4799250362861323dd059970a6408f5cdad64579211e95e7dc88daf2b0dbdb57137f287ea617148690f50e0e25899cdcdd35287a5e6c2341224d2b2af0b56c272022d8c7631f2364e2227ded03db547e985329bbc4d87c7243f90fc216f8294574f89fada55b1388ccc0fffcb69f6b04d0c7c5f81e3be64086c7abb8a772a1207ceed96d3ccd112fa610324ee84e7713d34d6c611dfffd744f0bd55c91887fb983a48c94b8531d6ec19508f83f311864666731c58082ebf5421f26c7e173116e95734074d7ca1dffa08d0ab79f97261f061d7d30765260b1be9fbbe98ea7b70014e0d53ec95c02cce857c357fd78e74c1f0ac508ab2e52d7f4d667e83606242317c83a2d93e129c2d8995f51c1963120135773729b222710039be54b36c4e8712c738960f1ca2845e6e27f252ff1dea64fa16469f4e4a34a496a16a5c9cc631137df53421473281a71955ad8db4fabac467cc8d3feac1391e9ed2dbf4a67a1464ef3dd8c0e6104db8df351ebeafffff061518110ced4ace0bf17bc8ddc58825f84e4d56617e30171a738d99b25018ca8f978f483095257762c4fa71ba530933acacc152429be2f4cfd712076252e9f2859e4ce0c95a92b8ad288586ed4aa307e1040be09f1ec23295f5f09f6873798eb331408bb394aac2494dc3694d8fcdeeb2fb72aaf96172d6a6108378362c733e25a2e1fa1e5545bbfa38c14e8255523456f7b85000ebe152620bc02d1b5838e72017b493519ebdcdf1a81974726b8fb3b5096af4138571940614ca87d73b4afc4d802585add4360862fa052fc50e9096b7bea3a83f0fe14f6e96b889dfa9d61789b9ef597d27ffefe7d1b23621a9eff06429eaeeb7efd28ee5618c7565b0964bb3c7d3222f957dc76103533be35f9558264fd93e6a0a40d97d1418532a7b0f30832eb9ea02a6035f702a0cfe9b7001f84f0d0f9bbd0d225258a18106b8dd4013386c033ad27231545542f4d735ec82f9722ed5aa80403172dd4146ea147ff3ddcc23427e6699d952468a6182d9e1f98dbc5945769f26b014d4883a91c47f7d3f76dcea392705f38d89dc369d53576a5b2baa5fe02bc0410";
    let params_raw = hex::decode(params_hex).unwrap();

    let params = ParamsKZG::<Bn256>::read(&mut BufReader::new(params_raw.as_slice()))
        .expect("restore params error");

    // TODO replace your circuit struct
    let empty_circuit = SimpleCircuit {
        constant: Fr::from(4),
        a: Value::unknown(),
        b: Value::unknown(),
    };

    match opts.sub {
        Subcommands::Solidity { file } => {
            let sol_code = gen_sol_verifier(&params, empty_circuit, vec![3])
                .expect("generate solidity file error");
            println!(
                "Generated verifier contract size: {}",
                evm::compile_solidity(sol_code.as_str()).len()
            );
            fs::write(file, sol_code).expect("write verifier solidity error");
        }

        Subcommands::Proof { private_a, private_b, project_id, task_id } => {
            let private_a = Fr::from(private_a);
            let private_b = Fr::from(private_b);
        
            let constant = Fr::from(4);

            let pk = gen_pk(&params, &empty_circuit);

            // TODO instance circuit
            let circuit = SimpleCircuit {
                constant,
                a: Value::known(private_a),
                b: Value::known(private_b),
            };
            // TODO public info
            let c = constant * private_a.square() * private_b.square();
            let instances = vec![vec![c, Fr::from(project_id), Fr::from(task_id)]];
        
            let proof = gen_proof(&params, &pk, circuit.clone(), &instances);
        
            println!("the proof is {:?}", hex::encode(&proof))
        }

        Subcommands::Verfiy { proof, public, project, task } => {
            let proof_raw = fs::read(proof).expect("read proof file error");
            let proof_raw = String::from_utf8(proof_raw).unwrap();
            let proof: Proof = serde_json::from_str(&proof_raw).unwrap();
            let proof = proof.proof;
            let proof = hex::decode(&proof[2..]).unwrap();

            let pk = gen_pk(&params, &empty_circuit);

            let instances = vec![vec![Fr::from(public), Fr::from(project), Fr::from(task)]];

            let calldata = encode_calldata(&instances, &proof);
            let accept = {
                let instances = instances
                    .iter()
                    .map(|instances| instances.as_slice())
                    .collect_vec();
                let mut transcript = TranscriptReadBuffer::<_, G1Affine, _>::init(proof.as_slice());
                VerificationStrategy::<_, VerifierGWC<_>>::finalize(
                    verify_proof::<_, VerifierGWC<_>, _, EvmTranscript<_, _, _, _>, _>(
                        params.verifier_params(),
                        pk.get_vk(),
                        AccumulatorStrategy::new(params.verifier_params()),
                        &[instances.as_slice()],
                        &mut transcript,
                    )
                    .unwrap(),
                )
            };

            println!("the proof is {:?}", accept);
            println!("calldata is {:?}", hex::encode(&calldata))
        }
    }
}
