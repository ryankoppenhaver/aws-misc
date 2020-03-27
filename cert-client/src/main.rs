use rusoto_core::{Region};
//use rusoto_s3::{S3Client, S3};
use rusoto_lambda::{LambdaClient,Lambda,InvocationRequest};
use bytes::Bytes;
//use std::env;
use std::fs::File;
use std::io::Read;

#[macro_use]
extern crate error_chain;

mod errors {
    error_chain! {}
}
use errors::*;

//const BUCKET_KEY: &str = "SSH_CERT_BUCKET";

fn get_arn() -> Result<String> {
    let mut f = File::open("/etc/config.json").chain_err(|| "open config")?;

    let mut s = String::new();
    f.read_to_string(&mut s).chain_err(|| "read config")?;

    let json = serde_json::from_str(&s).chain_err(|| "parse config")?;

    Ok("test".to_string())
}

fn main() -> Result<()> {
    let arn = get_arn()?;
    invoke(arn);
    Ok(())
}

/*
    match  {
        Err(e) => { eprintln!("open file: {:?}", e); return },
        Ok(mut f) => {
            let mut s = String::new();
            match f.read_to_string(&mut s) {
                Err(e) => { eprintln!("read file: {}", e); return },
                Ok(_) => {
                    match serde_json::from_str(&s) {
                        Err(e) => { eprintln!("parse json: {}", e); return },
                        Ok(arn) => invoke(arn)
                    }
                }
            }
        }
    }
}
*/

fn invoke(arn: String) { //-> Result<TKTK> {


    let lc = LambdaClient::new(Region::default()); //XXX TODO env not set on instance

    let payload_bytes = Bytes::from(&b"{\"hello\":\"world\"}"[..]);

    let resp = lc.invoke(InvocationRequest {
        function_name: arn,
        payload: Some(payload_bytes),

        ..Default::default()
    }).sync().expect("invoke failed");

    println!("{:?}", resp);

}

    /*
    if let Ok(bucket) = env::var(BUCKET_KEY) {
        println!("{}", bucket);
    } else {
        eprintln!("Missing env var {}", BUCKET_KEY)
    }*/



/*
    let s3 = S3Client::new(Region::UsWest2);

    let output = s3.list_buckets().sync().expect("list buckets:");
    output.buckets
        .unwrap_or_default()
        .iter().for_each(|b| println!("{}", b.name.as_deref().unwrap_or_default()));*/
