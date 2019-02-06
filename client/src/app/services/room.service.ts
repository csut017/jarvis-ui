import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { tap, catchError, map } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class RoomService {

  constructor(private http: HttpClient,
    private logging: LoggingService) {
    this.logger = logging.get('RoomService');
    this.current = new RoomInterface(http, logging, 'Test-1');
  }

  current: RoomInterface;
  private logger: Logger;

  list(): Observable<RoomInterface[]> {
    const url = `${environment.apiURL}rooms`;
    this.logger.log(`Retrieving rooms`, url);
    return this.http.get<roomList>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved rooms', res)),
        map(res => res.items.map(item => new RoomInterface(this.http, this.logging, item))),
        catchError(this.logger.handleError<RoomInterface[]>(`list()`))
      );
  }
}

export class RoomInterface {
  constructor(private http: HttpClient,
    logging: LoggingService,
    public name: string) {
    this.logger = logging.get('RoomInterface');
    this.uriName = encodeURIComponent(this.name);
  }

  private uriName: string;
  private logger: Logger;

  get(): Observable<RoomDetails> {
    const url = `${environment.apiURL}rooms/${this.uriName}`;
    this.logger.log(`Retrieving room details for ${this.name}`, url);
    return this.http.get<RoomDetails>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved room details', res)),
        catchError(this.logger.handleError<RoomDetails>(`'${this.name}'->get()`))
      );
  }
}

export interface RoomDetails {
}

interface roomList {
  items: string[];
}