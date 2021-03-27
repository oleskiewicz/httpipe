use std::io;
use std::path::PathBuf;

use async_process::{Command, Stdio};
use async_std::fs::read_to_string;
use futures_lite::io::{AsyncReadExt, AsyncWriteExt};
use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Method, Request, Response, Server, StatusCode};


fn error(status_code: StatusCode) -> Response<Body> {
    Response::builder()
        .status(status_code)
        .body("Not found".into())
        .unwrap()
}

async fn file_read(filename: &str) -> hyper::Result<Response<Body>> {
    if let Ok(s) = read_to_string(filename).await {
        let body = Body::from(s);
        return Ok(Response::new(body));
    }
    return Ok(error(StatusCode::NOT_FOUND))
}

fn get_file(name_fn: &str, name_file: &str) -> Result<PathBuf, std::io::Error> {
    // FIXME: filepath.push has a BS "replace if absolute" behaviour
    //        THIS CAN CAUSE RUNTIME PANIC, hence the check
    if name_fn.len() < 1 {
        return Err(io::Error::new(io::ErrorKind::InvalidInput, "path too short"));
    }
    let mut filepath = PathBuf::from( "./fn/");
    filepath.push(&name_fn[1..]);

    if filepath.is_absolute() {
        // FIXME: pass errors up the call stack
        return Err(io::Error::new(io::ErrorKind::InvalidInput, "wrong path"));
    }
    filepath.push(name_file);

    if !filepath.is_file() {
        return Err(io::Error::new(io::ErrorKind::NotFound, "file not found"));
    }

    Ok(filepath)
}

async fn run(req: Request<Body>) -> hyper::Result<Response<Body>> {
    match (req.method(), req.uri().path()) {

        (&Method::GET, "/") => {
            if let Ok(doc) = get_file("/", "doc") {
                return file_read(&doc.to_str().unwrap()).await;
            }
            Ok(error(StatusCode::NOT_FOUND))
        },

        (&Method::GET, name) => {
            if let Ok(doc) = get_file(name, "doc") {
                return file_read(&doc.to_str().unwrap()).await;
            }
            Ok(error(StatusCode::NOT_FOUND))
        },

        (&Method::POST, name) => {
            if let Ok(handler) = get_file(name, "handle") {
                let body = hyper::body::to_bytes(req.into_body()).await?;
                let child = Command::new(handler)
                    .env_clear()
                    .stdin(Stdio::piped())
                    .stdout(Stdio::piped())
                    .spawn()
                    .expect("Failed to spawn child process");

                let mut out = Vec::new();
                let _ = child.stdin.unwrap().write_all(&body).await;
                // FIXME: reading a vec regardless of mimetype might be bad
                let _ = child.stdout.unwrap().read_to_end(&mut out).await.unwrap();

                let res = Response::new(Body::from(out));
                return Ok(res);
            }
            Ok(error(StatusCode::NOT_FOUND))
        }

        _ => {
            Ok(error(StatusCode::NOT_FOUND))
        }

    }
}

#[tokio::main]
pub async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let addr = ([127, 0, 0, 1], 3000).into();
    let service = make_service_fn(|_| async { Ok::<_, hyper::Error>(service_fn(run)) });
    let server = Server::bind(&addr).serve(service);
    println!("Listening on http://{}", addr);
    server.await?;
    Ok(())
}
